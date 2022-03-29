StreamingFast dbin Library
------------------
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/dbin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`dbin` is a simple file storage format to pack a stream of protobuf
messages. It is part of **[StreamingFast](https://github.com/streamingfast/streamingfast)**.

## Usage

See example usage in [merger](https://github.com/streamingfast/merger)


## Latest file format (v1)

First four magic bytes:
* 'd', 'b', 'i', 'n'

Next single byte:
* file format version, current is `0x01`

Next 2 bytes are big-endian uint16, the length of the proto definition type, to follow:
* 0x1900

Next bytes are the name of the proto def:
* "sf.ethereum.type.v1.Block"

Rest of the file, a sequence of:
* Length-prefixed messages, with each length specified as 4 bytes big-endian uint32.
* Followed by message of that length

EOF reached when no more bytes exist after the last message boundary.


## Prior file format (v0)

Version 0 of the `dbin` for mat was:

First four magic bytes:
* 'd', 'b', 'i', 'n'

Next single byte:
* file format version, was `0x00`.

Next 5 bytes were the content type, including 3 letters and 2 numbers.

Rest of the file, a sequence of:
* Length-prefixed messages, with each length specified as 4 bytes big-endian uint32.
* Followed by message of that length

EOF reached when no more bytes exist after the last message boundary.


## Contributing

**Issues and PR in this repo related strictly to the dbin library.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.


## License

[Apache 2.0](LICENSE)
