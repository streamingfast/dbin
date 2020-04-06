package dbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
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

func (r *Reader) ReadHeader() (contentType string, version int32, err error) {
	if r.readHeaderDone {
		return "", 0, fmt.Errorf("header was read already")
	}

	header, err := r.readCompleteBytes(10, "header")
	if err != nil {
		return "", 0, err
	}

	r.readHeaderDone = true
	if !bytes.HasPrefix(header, magicString) {
		return "", 0, fmt.Errorf("magic string 'dbin' not found in header")
	}

	if header[4] != fileVersion {
		return "", 0, fmt.Errorf("invalid dbin file revision, expected %d, got %d", fileVersion, header[4])
	}

	contentType := string(header[5:8])
	version, err := strconv.ParseInt(string(header[8:10]), 10, 32)
	if err != nil {
		return "", 0, err
	}

	return contentType, int32(version), nil
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

	messageBytes, err := r.readCompleteBytes(length, "message")
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

func (r *Reader) readCompleteBytes(length int, tag string) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := io.ReadFull(r.Reader, bytes)

	return bytes, err
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
