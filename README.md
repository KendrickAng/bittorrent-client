# BitTorrent Client (with 🧲 support!)

A command-line BitTorrent client
implementing [BEP 3: The BitTorrent Protocol Specification](https://www.bittorrent.org/beps/bep_0003.html) (torrent file
support)
and [BEP 9: Extension for Peers to Send Metadata Files](https://www.bittorrent.org/beps/bep_0009.html) (magnet link
support).

[![asciicast](https://asciinema.org/a/u7JXu5EJPGialBWua7jyKXajN.svg)](https://asciinema.org/a/u7JXu5EJPGialBWua7jyKXajN)

>💡 Interested in how this works under the hood? Check out the [wiki](https://kendrickang.github.io/bittorrent-client-with-magnet/)!

## Quickstart

First build the `btclient` binary:

```shell
go build
```

Next, either download a torrent starting from a magnet link:

```shell
./btclient -type=magnet sample.magnet
```

Or download a torrent starting from a `.torrent` file:

```shell
./btclient -type=torrent sample.torrent
```

## Credits

[CodeCrafters](https://app.codecrafters.io/courses/bittorrent/overview) for their sample `.torrent` and `.magnet` files.
