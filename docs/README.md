# BitTorrent Client (with 🧲) Wiki

<script src="https://asciinema.org/a/u7JXu5EJPGialBWua7jyKXajN.js" id="asciicast-u7JXu5EJPGialBWua7jyKXajN" async="true"></script>

This is the wiki page for the BitTorrent client
implementing [BEP 3: The BitTorrent Protocol Specification](https://www.bittorrent.org/beps/bep_0003.html) and [BEP 9: Extension for Peers to Send Metadata Files](https://www.bittorrent.org/beps/bep_0009.html).

After reading this wiki, you should understand the various components that comprise this codebase, reason about how they interact with one another, and ultimately be able to adapt this code to your needs or write a BitTorrent client yourself 🙂.

Let's get started!

## Overview

Before we begin, I recommended skimming the following in order:

1. Jesse Li's [Building a BitTorrent client from the ground-up in Go](https://blog.jse.li/posts/torrent/)
2. [BEP 3: The BitTorrent Protocol Specification](https://www.bittorrent.org/beps/bep_0003.html)
3. [BEP 9: Extension for Peers to Send Metadata Files](https://www.bittorrent.org/beps/bep_0009.html)

You should have an idea of how to answer the following:

- What are the high-level steps a torrent client takes? (Hint: Parse torrent file, Query the Tracker, Connect to peers, Download from peers, Regenerate file from downloaded bytes)
- What information does the `.torrent` file expose?
- How does a peer connect to another peer and start exchanging information?
- What is the API between two peers (i.e. operations supported by the peer)?

## Architecture

>TODO

### Directory Structure

- `/`: Top-level directory containing `main.go` and flag parsing logic.
- `internal/`: Libraries used internally by the program.
  - `bittorrent/`: Libraries related to the BitTorrent protocol.
    - `client/`: Given a torrent file, coordinates downloads from peers. 
    - `handshake/`: Handles initial connection to a peer.
    - `message/`: Contains data structures for messages exchanged between peers.
    - `peer/`: Abstracts a connection to a single peer.
    - `torrentfile/`: Abstracts operations on the `.torrent` file.
    - `tracker/`: Abstracts operations between the client and a BitTorrent tracker.
  - `preconditions/`: Utility methods.
  - `stringutil/`: Utility methods.
  - `udpprotocol/`: Unused. Ignore this folder.

## Parse Torrent File

>TODO

## Query Tracker

>TODO

## Connect to Peers

>TODO

## Download from Peers

>TODO

## Regenerate file

>TODO

## Debugging

>TODO: Insert Wireshark debugging screenshot or video

[Wireshark](https://www.wireshark.org/) is a network packet analyzer that understands the format of packets specified by the BitTorrent protocol. Analyzing the packets sent or received (or lack thereof) will let you know if any packet is misconstructed.

## Future Work

- Allow the client to upload, not just download
