package bencodeutil

import (
	"errors"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
)

type TorrentMetainfo struct {
	// REQUIRED. The announce URL of the tracker.
	Announce string "announce"

	// REQUIRED. Dictionary describing files of the torrent. Can in 'single file' or 'multi file' format.
	Info Info "info"

	// TODO: Add support for this (https://www.bittorrent.org/beps/bep_0012.html).
	// OPTIONAL. Extension to the official specification for Announce.
	AnnounceList [][]string "announce-list"

	// OPTIONAL. Creation time of the torrent, in standard UNIX epoch format.
	CreationDate uint64 "creation date"

	// OPTIONAL. Free-form text comments of the author.
	Comment string "comment"

	// OPTIONAL. Name and version of the program used to create the .torrent.
	CreatedBy string "created by"

	// OPTIONAL. The string encoding format used to generate 'pieces' in the info dictionary.
	Encoding string "encoding"
}

type Info struct {
	// CONFIGURATIONS FOR BOTH SINGLE/MULTI FILE MODE.

	// REQUIRED. Number of bytes in each piece.
	PieceLength int "piece length"

	// REQUIRED. Concatenation of all 20-byte SHA1 hash value, one per piece.
	Pieces string "pieces"

	// TODO: Add support for this.
	// OPTIONAL. If '1', client get peers ONLY via trackers in the metainfo file.
	// If '0', or not present, client may obtain peer from other means, e.g. PEX peer exchange, dht.
	Private int "private"

	// REQUIRED. Single: The file name; purely advisory.
	// Multi: The name of the directory where files are stored; purely advisory.
	Name string "name"

	// CONFIGURATIONS FOR SINGLE FILE MODE.

	// REQUIRED. The length of the file in bytes.
	Length int "length"

	// OPTIONAL. A 32-character hex string corresponding to the MD5 sum of the file.
	MD5Sum string "md5sum"

	// CONFIGURATIONS FOR MULTI FILE MODE.

	// REQUIRED. Each file describes a file.
	Files []Files "files"
}

type Files struct {
	// REQUIRED. Length of the file in bytes.
	Length int "length"

	// REQUIRED. Filepath that forms the dir1/dir2/file.ext path when joined.
	Path []string "path"

	// OPTIONAL. 32-character hex string corresponding to the MD5 sum of the file.
	MD5Sum string "md5sum"
}

func Unmarshal(r io.Reader) (TorrentMetainfo, error) {
	var data TorrentMetainfo

	if err := bencode.Unmarshal(r, &data); err != nil {
		return TorrentMetainfo{}, errors.Join(err, fmt.Errorf("bencodeutil failed to unmarshal %+v", r))
	}

	return data, nil
}
