package dbin

import (
	"encoding/binary"
	"fmt"
	"io"
)

const fileVersion = byte(1)
const maxContentTypeLength = 65535

type Writer struct {
	io.Writer
	headerWritten bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

// WriteHeader writes the header of the file, it should be called only once.
// contentType ideally is the fullyQualified name of the protocol buffer message
func (w *Writer) WriteHeader(contentType string) error {
	if w.headerWritten {
		return fmt.Errorf("header already written")
	}

	cntType := []byte(contentType)
	if len(cntType) > maxContentTypeLength {
		return fmt.Errorf("contentType length should not exceed %d characters, was %d", maxContentTypeLength, len(cntType))
	}

	if len(cntType) == 0 {
		return fmt.Errorf("contentType should contain at-least one character")
	}

	header := []byte{'d', 'b', 'i', 'n', fileVersion, 0x00, 0x00}
	binary.BigEndian.PutUint16(header[5:], uint16(len(contentType)))
	header = append(header, []byte(contentType)...)
	bytesWritten, err := w.Write(header)
	if err != nil {
		return fmt.Errorf("unable to write header: %s", err)
	}

	expectedBytesWritten := 7 + len(contentType)

	if bytesWritten != expectedBytesWritten {
		return fmt.Errorf("incomplete header write, expected %d bytes written: wrote only %d bytes", expectedBytesWritten, bytesWritten)
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
