package torrentfile

import (
	"bytes"
	"errors"
	"example.com/btclient/pkg/bittorrent"
	"example.com/btclient/pkg/stringutil"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
)

// TorrentFile represents a decoded Metainfo (.torrent) file which was originally bencoded.
type TorrentFile struct {
	// REQUIRED. The announce URL of the tracker.
	Announce string `bencode:"announce"`

	// REQUIRED. Dictionary describing files of the torrent. Can in 'single file' or 'multi file' format.
	Info Info `bencode:"info"`

	// TODO: Add support for this (https://www.bittorrent.org/beps/bep_0012.html).
	// OPTIONAL. Extension to the official specification for Announce.
	AnnounceList [][]string `bencode:"announce-list"`

	// OPTIONAL. Creation time of the torrent, in standard UNIX epoch format.
	CreationDate uint64 `bencode:"creation date"`

	// OPTIONAL. Free-form text comments of the author.
	Comment string `bencode:"comment"`

	// OPTIONAL. Name and version of the program used to create the .torrent.
	CreatedBy string `bencode:"created by"`

	// OPTIONAL. The string encoding format used to generate 'pieces' in the info dictionary.
	Encoding string `bencode:"encoding"`
}

// Info represents the info dictionary containing information about the torrent data.
type Info struct {
	// CONFIGURATIONS FOR BOTH SINGLE/MULTI FILE MODE.

	// REQUIRED. Number of bytes in each piece.
	PieceLength int `bencode:"piece length,omitempty"`

	// REQUIRED. Concatenation of all 20-byte SHA1 hash value, one per piece.
	Pieces string `bencode:"pieces,omitempty"`

	// TODO: Add support for this.
	// OPTIONAL. If '1', client get peers ONLY via trackers in the metainfo file.
	// If '0', or not present, client may obtain peer from other means, e.g. PEX peer exchange, dht.
	Private int `bencode:"private,omitempty"`

	// REQUIRED. Single: The file name; purely advisory.
	// Multi: The name of the directory where files are stored; purely advisory.
	Name string `bencode:"name,omitempty"`

	// CONFIGURATIONS FOR SINGLE FILE MODE.

	// REQUIRED. The length of the file in bytes.
	Length int `bencode:"length,omitempty"`

	// OPTIONAL. A 32-character hex string corresponding to the MD5 sum of the file.
	MD5Sum string `bencode:"md5sum,omitempty"`

	// CONFIGURATIONS FOR MULTI FILE MODE.

	// REQUIRED. Each file describes a file.
	Files []Files `bencode:"files,omitempty"`
}

// Files represents a set of files that go in a directory structure.
type Files struct {
	// REQUIRED. Length of the file in bytes.
	Length int `bencode:"length"`

	// REQUIRED. Filepath that forms the dir1/dir2/file.ext path when joined.
	Path []string `bencode:"path"`

	// OPTIONAL. 32-character hex string corresponding to the MD5 sum of the file.
	MD5Sum string `bencode:"md5sum"`
}

// ReadTorrentFile reads and returns a [TorrentFile] from r.
func ReadTorrentFile(r io.Reader) (TorrentFile, error) {
	var data TorrentFile

	if err := bencode.Unmarshal(r, &data); err != nil {
		return TorrentFile{}, errors.Join(err, fmt.Errorf("bencodeutil failed to unmarshal %+v", r))
	}

	if err := data.Validate(); err != nil {
		return TorrentFile{}, err
	}

	return data, nil
}

// Simplify flattens [TorrentFile] and returns a [SimpleTorrentFile].
func (t *TorrentFile) Simplify() (SimpleTorrentFile, error) {
	// SHA-1 hash of info dict
	buf := new(bytes.Buffer)
	if err := bencode.Marshal(buf, t.Info); err != nil {
		return SimpleTorrentFile{}, err
	}
	bufHash := bittorrent.Hash(buf.Bytes())

	// Split pieces into pieces of 20 bytes each
	sha1Chunks, err := stringutil.SplitChunksOf20(t.Info.Pieces)
	if err != nil {
		return SimpleTorrentFile{}, err
	}

	return SimpleTorrentFile{
		Announce:    t.Announce,
		InfoHash:    bufHash,
		PieceHashes: sha1Chunks,
		PieceLength: t.Info.PieceLength,
		Name:        t.Info.Name,
		Length:      t.Info.Length,
	}, nil
}

// Validate performs nonblocking validations on [TorrentFile].
func (t *TorrentFile) Validate() error {
	if t.Announce == "" {
		return errors.New("torrent file has no announce")
	}
	if t.Info.PieceLength <= 0 {
		return fmt.Errorf("invalid piece length: %d", t.Info.PieceLength)
	}
	if len(t.Info.Pieces) == 0 {
		return errors.New("torrent file has no pieces")
	}
	if (len(t.Info.Pieces) % 20) != 0 {
		return fmt.Errorf("expected pieces length to be multiple of 20, got %d", len(t.Info.Pieces))
	}
	return nil
}
