package dbin

import (
	"encoding/binary"
	"fmt"
	"io"
)

const fileVersion = byte(0)

type Writer struct {
	io.Writer
	headerWritten bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

func (w *Writer) WriteHeader(contentType string, version int) error {
	if w.headerWritten {
		return fmt.Errorf("header already written")
	}

	cntType := []byte(contentType)
	if len(cntType) != 3 {
		return fmt.Errorf("contentType should be 3 characters, was %d %v", len(cntType), cntType)
	}
	if version > 99 || version < 0 {
		return fmt.Errorf("version should be between 0 and 99, was %d", version)
	}

	ver := []byte(fmt.Sprintf("%02d", version))

	written, err := w.Write([]byte{'d', 'b', 'i', 'n', fileVersion, cntType[0], cntType[1], cntType[2], ver[0], ver[1]})
	if err != nil {
		return err
	}

	w.headerWritten = true

	if written != 10 {
		return fmt.Errorf("incomplete header write (10 bytes): wrote only %d bytes", written)
	}

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
