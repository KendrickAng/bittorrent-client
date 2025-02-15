# BitTorrent Client

A command-line BitTorrent client
implementing [BEP 3: The BitTorrent Protocol Specification](https://www.bittorrent.org/beps/bep_0003.html) (torrent file
support)
and [BEP 9: Extension for Peers to Send Metadata Files](https://www.bittorrent.org/beps/bep_0009.html) (magnet link
support).

## Quickstart

Run with a magnet link:

```shell
go build && ./btclient -type=magnet sample.magnet
```

Run with a torrent file:

```shell
go build && ./btclient -type=torrent sample.torrent
```

## Credits

[CodeCrafters](https://app.codecrafters.io/courses/bittorrent/overview) for their sample `.torrent` and `.magnet` files.

## Future Work
- Write a script for cross-platform compilation and running
