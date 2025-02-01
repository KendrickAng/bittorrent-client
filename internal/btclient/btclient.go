package btclient

import (
	"context"
	"errors"
	"example.com/btclient/pkg/bittorrent"
	"example.com/btclient/pkg/bittorrent/client"
	"example.com/btclient/pkg/bittorrent/handshake"
	"example.com/btclient/pkg/bittorrent/peer"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"example.com/btclient/pkg/bittorrent/tracker"
	"example.com/btclient/pkg/preconditions"
	"example.com/btclient/pkg/stringutil"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Wait until SIGINT is given, or the handler succeeds
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		defer cancel()
		<-signals
	}()

	// Parse flags
	flags, err := GetFlags()
	if err != nil {
		return err
	}

	preconditions.CheckArgument(preconditions.Xor(flags.TorrentFileName != "", flags.MagnetFileName != ""),
		"only either torrent file or magnet link must be specified")

	if flags.TorrentFileName != "" {
		return runWithTorrentFile(ctx, flags.TorrentFileName)
	}

	if flags.MagnetFileName != "" {
		return runWithMagnet(ctx, flags.MagnetFileName)
	}

	return nil
}

func runWithTorrentFile(ctx context.Context, torrentFileName string) (err error) {
	// Read .torrent file
	file, err := os.Open(torrentFileName)
	if err != nil {
		return err
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// Decode bencoded file
	bencodedData, err := torrentfile.ReadTorrentFile(file)
	if err != nil {
		return err
	}
	torrent, err := bencodedData.Simplify()
	if err != nil {
		return err
	}

	// Parse tracker response
	trackerResp, err := tracker.DefaultHttpClient.FetchTorrentMetadata(tracker.FetchTorrentMetadataRequest{
		TrackerUrl: torrent.Announce,
		InfoHash:   torrent.InfoHash,
		PeerID:     torrent.InfoHash,
		Left:       torrent.Length,
	})
	if err != nil {
		return err
	} else if len(trackerResp.Peers) == 0 {
		return errors.New("no peers found")
	} else {
		torrent.Peers = trackerResp.Peers
		println("parsed tracker response")
	}

	extensionBits := bittorrent.NewExtensionBits(bittorrent.ExtensionProtocolBit)
	clients, err := connectToClients(trackerResp.Peers, extensionBits, torrent.PeerID, torrent.InfoHash)
	if err != nil {
		return err
	}
	connectionPool := peer.NewPool(clients)

	// Handle (blocking)
	handler, err := client.NewClient(torrent, connectionPool)
	if err != nil {
		return err
	}
	if _, err := handler.Handle(ctx); err != nil {
		return err
	}
	defer handler.Close()

	return nil
}

func runWithMagnet(ctx context.Context, magnetFileName string) (err error) {
	file, err := os.Open(magnetFileName)
	if err != nil {
		return err
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// Read magnet link from the file.
	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Parse magnet link.
	mag, err := bittorrent.ParseMagnet(string(b))
	if err != nil {
		return err
	}

	// Create peer ID.
	peerID, err := stringutil.Random20Bytes()
	if err != nil {
		return err
	}

	// Download tracker information.
	infoHash, err := mag.InfoHash()
	if err != nil {
		return err
	}

	var trackerResp *tracker.Response
	for _, trackerUrl := range mag.TrackerUrls() {
		trackerResp, err = tracker.DefaultHttpClient.FetchTorrentMetadata(tracker.FetchTorrentMetadataRequest{
			TrackerUrl: trackerUrl,
			InfoHash:   infoHash,
			PeerID:     peerID,
			Left:       999, // we don't know the file size in advance; use a made-up value as workaround
		})
		if err != nil {
			continue
		} else {
			break
		}
	}
	if trackerResp == nil {
		return errors.New("could not retrieve tracker information")
	}

	// Connect to clients.
	extensionBits := bittorrent.NewExtensionBits(bittorrent.ExtensionProtocolBit)
	clients, err := connectToClients(trackerResp.Peers, extensionBits, peerID, infoHash)
	if err != nil {
		return err
	}

	// Retrieve info dict from any peer
	var infoDict *torrentfile.Info
	for _, peerClient := range clients {
		if peerClient.InfoDict != nil {
			infoDict = peerClient.InfoDict
			break
		}
	}

	if infoDict == nil {
		return errors.New("could not retrieve info dictionary")
	}

	// Convert info dict into a torrent file representation
	torrentFile := torrentfile.TorrentFile{
		PeerId: peerID,
		Info:   *infoDict,
	}
	simpleTorrentFile, err := torrentFile.Simplify()
	if err != nil {
		return err
	}

	// Handle (blocking)
	connectionPool := peer.NewPool(clients)
	handler, err := client.NewClient(simpleTorrentFile, connectionPool)
	if err != nil {
		return err
	}
	if _, err := handler.Handle(ctx); err != nil {
		return err
	}
	defer handler.Close()

	return nil
}

func connectToClients(peers []netip.AddrPort,
	extension bittorrent.ExtensionBits,
	peerID [20]byte,
	infoHash [20]byte) ([]*peer.Client, error) {

	peerClientCh := make(chan *peer.Client, len(peers))
	wg := new(sync.WaitGroup)
	for _, addrPort := range peers {
		wg.Add(1)

		go func(toConnect netip.AddrPort) {
			defer wg.Done()

			peerClient, err := connectToClient(toConnect, extension, peerID, infoHash)
			if err != nil {
				println("error creating client for peer", toConnect.String(), err.Error())
				return
			}

			peerClientCh <- peerClient
			fmt.Printf("created client for %s\n", toConnect.String())
		}(addrPort)
	}
	wg.Wait()
	close(peerClientCh) // close channel so we don't loop over it infinitely

	// Convert peers channel into peers queue
	var clients []*peer.Client
	for peerClient := range peerClientCh {
		clients = append(clients, peerClient)
	}
	fmt.Printf("found %d peers\n", len(clients))

	return clients, nil
}

func connectToClient(addrPort netip.AddrPort,
	ext bittorrent.ExtensionBits,
	peerID [20]byte,
	infoHash [20]byte) (*peer.Client, error) {

	// dial peer
	conn, err := net.DialTimeout("tcp", addrPort.String(), 30*time.Second)
	if err != nil {
		return nil, err
	}
	println("dialed", conn.RemoteAddr().String())

	// create client to peer
	peerClient := peer.NewClient(conn,
		conn,
		handshake.NewHandshaker(conn),
		ext,
		peerID,
		infoHash)
	if err := peerClient.Init(); err != nil {
		return nil, err
	}

	return peerClient, err
}
