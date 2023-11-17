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

func (r *Reader) ReadHeader() (*Header, error) {
	if r.readHeaderDone {
		return nil, fmt.Errorf("header was read already")
	}

	partialHeader, err := r.readBytes(5)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(partialHeader, magicString) {
		return nil, fmt.Errorf("magic string 'dbin' not found in header")
	}

	ver := partialHeader[4]
	header := &Header{
		RawBytes:    partialHeader,
		Version:     ver,
		ContentType: "",
	}

	if ver == 0 {
		contentTypeBytes, err := r.readBytes(3)
		if err != nil {
			return nil, fmt.Errorf("failed to read content type: %s", err)
		}
		header.RawBytes = append(header.RawBytes, contentTypeBytes...)
		header.ContentType = string(contentTypeBytes)

		dataVersionByte, err := r.readBytes(2)
		if err != nil {
			return nil, fmt.Errorf("failed to read content version: %w", err)
		}
		header.RawBytes = append(header.RawBytes, dataVersionByte...)
		r.readHeaderDone = true
		return header, nil
	}

	if ver == 1 {
		contentTypeLength, err := r.readBytes(2)
		if err != nil {
			return nil, fmt.Errorf("reading content type length: %w", err)
		}
		header.RawBytes = append(header.RawBytes, contentTypeLength...)

		length := binary.BigEndian.Uint16(contentTypeLength)
		contentTypeBytes, err := r.readBytes(int(length))
		if err != nil {
			return nil, fmt.Errorf("reading content type of length %d: %w", length, err)
		}
		header.RawBytes = append(header.RawBytes, contentTypeBytes...)
		header.ContentType = string(contentTypeBytes)
		r.readHeaderDone = true
		return header, nil
	}
	return nil, fmt.Errorf("unsupported file format version: %d", ver)
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
