StreamingFast dbin Library
------------------
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/dbin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`dbin` is a simple file storage format to pack a stream of protobuf
messages. It is part of **[StreamingFast](https://github.com/streamingfast/streamingfast)**.

## Usage

See example usage in [merger](https://github.com/streamingfast/merger)


## File format

First four magic bytes:
* 'd', 'b', 'i', 'n'

Next single byte:
* file format version, current is `0x00`

Next three bytes:
* content type, like 'ETH', 'EOS', or whatever..

Next two bytes:
* 10-based string representation of content version: '00' for version 0, '99', for version 99

Rest of the file:
* Length-prefixed messages, with each length specified as 4 bytes big-endian uint32.
* Followed by message of that length, then start over.
* EOF reached when no more bytes exist after the last message boundary.


## Contributing

**Issues and PR in this repo related strictly to the dbin library.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.


## License

[Apache 2.0](LICENSE)
