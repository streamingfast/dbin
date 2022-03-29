package dbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var magicString = []byte("dbin")

type Reader struct {
	io.Reader
	readHeaderDone bool
}

func NewFileReader(filename string) (*Reader, error) {
	fl, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewReader(fl), nil
}

func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: r}
}

func (r *Reader) ReadHeader() (contentType string, err error) {
	if r.readHeaderDone {
		return "", fmt.Errorf("header was read already")
	}

	header, err := r.readBytes(5)
	if err != nil {
		return "", fmt.Errorf("reading header: %w", err)
	}

	if !bytes.HasPrefix(header, magicString) {
		return "", fmt.Errorf("magic string 'dbin' not found in header")
	}

	fmt.Println("Header", header)
	ver := header[4]
	switch ver {
	case 0:
		contentType, err := r.readBytes(5)
		if err != nil {
			return "", fmt.Errorf("reading content type of file v0: %w", err)
		}
		r.readHeaderDone = true
		return string(contentType), nil

	case 1:
		contentTypeLength, err := r.readBytes(2)
		if err != nil {
			return "", fmt.Errorf("reading content type length: %w", err)
		}
		length := binary.BigEndian.Uint16(contentTypeLength)
		contentType, err := r.readBytes(int(length))
		if err != nil {
			return "", fmt.Errorf("reading content type of length %d: %w", length, err)
		}
		r.readHeaderDone = true
		return string(contentType), nil
	}
	return "", fmt.Errorf("invalid dbin file version, expected 0 or 1, got %d", ver)
}

// ReadMessage reads next message from byte stream
func (r *Reader) ReadMessage() ([]byte, error) {
	lengthBytes, err := r.readBytes(4)
	if err == io.EOF {
		return nil, err
	}

	if len(lengthBytes) < 4 {
		return nil, incompleteReadError(err, "incomplete message length required %d bytes, got %d bytes", 4, len(lengthBytes))
	}

	length := int(binary.BigEndian.Uint32(lengthBytes))
	if length == 0 {
		return []byte{}, err
	}

	messageBytes, err := r.readBytes(length)
	if messageBytes == nil {
		return nil, err
	}

	return messageBytes, err
}

func (r *Reader) Close() error {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func (r *Reader) readBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := io.ReadFull(r.Reader, bytes)

	return bytes, err
}

func incompleteReadError(err error, message string, args ...interface{}) error {
	value := fmt.Sprintf(message, args...)
	if err != nil {
		value += ": " + err.Error()
	}

	return errors.New(value)
}
