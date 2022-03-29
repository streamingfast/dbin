package dbin

// reference
//First four bytes:
//* 'd', 'b', 'i', 'n'
//
//Next single byte:
//* file format version, current is `0x00`
//
//Next three bytes:
//* content type, like 'ETH', 'EOS', or whatever..
//
//Next two bytes:
//* 10-based string representation of content version: '00' for version 0, '99', for version 99
//
//Rest of the file:
//* Length-prefixed messages, with each length specified as 4 bytes big-endian uint32.
//* Followed by message of that length, then start over.
//* EOF reached when no more bytes exist after the last message boundary.

import (
	"encoding/binary"
	"fmt"
	"io"
)

const fileVersion = byte(1)

type Writer struct {
	io.Writer
	headerWritten bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

func (w *Writer) WriteHeader(contentType string) error {
	if w.headerWritten {
		return fmt.Errorf("header already written")
	}
	if len(contentType) > 65535 {
		return fmt.Errorf("content type too long, expected maximum 65535 in length, found %d bytes", len(contentType))
	}

	header := []byte{'d', 'b', 'i', 'n', fileVersion, 0x00, 0x00}
	binary.BigEndian.PutUint16(header[5:], uint16(len(contentType)))
	header = append(header, []byte(contentType)...)

	written, err := w.Write(header)
	if err != nil {
		return err
	}

	expectedWrite := 7 + len(contentType)
	if written != expectedWrite {
		return fmt.Errorf("expected %d bytes written, wrote %d", expectedWrite, written)
	}

	w.headerWritten = true

	return nil
}

func (w *Writer) WriteMessage(msg []byte) error {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(msg)))
	written, err := w.Write(length)
	if err != nil {
		return err
	}

	if written != 4 {
		return fmt.Errorf("incomplete length write (4 bytes): wrote %d bytes", written)
	}

	written, err = w.Write(msg)
	if err != nil {
		return err
	}

	if written != len(msg) {
		return fmt.Errorf("incomplete message write (%d bytes): wrote %d bytes", len(msg), written)
	}

	return nil
}

func (w *Writer) Close() error {
	if closer, ok := w.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
