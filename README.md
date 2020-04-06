`dbin` is a simple file storage format to pack a stream of protobuf messages

Today, you can say you played in `dbin`.


File format
~~~~~~~~~~~

First four bytes:
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
