package dbin

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHeader(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectType  string
		expectError string
	}{
		{
			name:       "happy path version 0",
			input:      []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
			expectType: "ETH",
		},
		{
			name:       "happy path version 1",
			input:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x2D, 't', 'y', 'p', 'e', '.', 'g', 'o', 'o', 'g', 'l', 'e', 'a', 'p', 'i', 's', '.', 'c', 'o', 'm', '/', 's', 'f', '.', 'e', 't', 'h', 'e', 'r', 'e', 'u', 'm', '.', 't', 'y', 'p', 'e', '.', 'v', '2', '.', 'B', 'l', 'o', 'c', 'k'},
			expectType: "type.googleapis.com/sf.ethereum.type.v2.Block",
		},
		{
			name:        "bad prefix v0",
			input:       []byte{'d', 'b', 'o', 'b', 0x00, 'E', 'T', 'H', '9', '8'},
			expectError: "magic string 'dbin' not found in header",
		},
		{
			name:        "bad prefix v1",
			input:       []byte{'d', 'b', '0', 'b', 0x01, 0x00, 0x2D, 't', 'y', 'p', 'e', '.', 'g', 'o', 'o', 'g', 'l', 'e', 'a', 'p', 'i', 's', '.', 'c', 'o', 'm', '/', 's', 'f', '.', 'e', 't', 'h', 'e', 'r', 'e', 'u', 'm', '.', 't', 'y', 'p', 'e', '.', 'v', '2', '.', 'B', 'l', 'o', 'c', 'k'},
			expectError: "magic string 'dbin' not found in header",
		},
		{
			name:        "file format version",
			input:       []byte{'d', 'b', 'i', 'n', 0x02, 'E', 'T', 'H', '9', '8'},
			expectError: "unsupported file format version: 2",
		},
		{
			name:        "incomplete header v0",
			input:       []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H'},
			expectError: "failed to read content version: EOF",
		},
		{
			name:        "incomplete header v1, content type doesn't correspond to length",
			input:       []byte{'d', 'b', 'i', 'n', 0x01, 0x2D, 'e', 't', 'h'},
			expectError: "reading content type of length 11621: unexpected EOF",
		},
		{
			name:        "incomplete header v1, missing content type length",
			input:       []byte{'d', 'b', 'i', 'n', 0x01, 0x2D},
			expectError: "reading content type length: unexpected EOF",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(test.input))

			header, err := r.ReadHeader()

			if test.expectError == "" {
				require.NoError(t, err)
				assert.Equal(t, test.expectType, header.ContentType)
				assert.Equal(t, test.input, header.Data)
			} else {
				require.Error(t, err)
				assert.Equal(t, test.expectError, err.Error())
			}
			assert.NoError(t, r.Close())
		})
	}
}

func TestDoubleReadHeader(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'}))
	r.ReadHeader()

	_, err := r.ReadHeader()

	assert.Error(t, err)
	assert.Equal(t, "header was read already", err.Error())
}

func TestReadMessageHappy(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{
		'd', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8',
		0x00, 0x00, 0x00, 0x01, // msg 1, length
		0x61,
	}))
	_, err := r.ReadHeader()
	require.NoError(t, err)

	message, err := r.ReadMessage()

	assert.NoError(t, err)
	assert.Equal(t, "a", string(message))

	_, err = r.ReadMessage()

	assert.Equal(t, io.EOF, err)
}

func TestReadMessage_OnlyHeader(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{
		'd', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8',
	}))
	_, err := r.ReadHeader()
	require.NoError(t, err)

	message, err := r.ReadMessage()
	assert.Nil(t, message)
	assert.Equal(t, io.EOF, err)
}

func TestReadMessage_IncompleteMessageLength(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00, 0x00},
	))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	_, err = r.ReadMessage()

	require.Error(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestReadMessage_MessageMissing(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{
		'd', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8',
		0x00, 0x00, 0x00, 0x01,
	}))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	_, err = r.ReadMessage()

	require.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

func TestReadMessage_MessageEmpty(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{
		'd', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8',
		0x00, 0x00, 0x00, 0x00,
	}))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	msg, err := r.ReadMessage()

	require.NoError(t, err)
	assert.Len(t, msg, 0)
}

func TestReadMessage_MesageLength_MultiReadNeeded(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00},
		[]byte{0x00, 0x00},
	))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	msg, err := r.ReadMessage()

	require.NoError(t, err)
	assert.Len(t, msg, 0)
}

func TestReadMessage_Mesage_MultiReadNeeded(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00, 0x00, 0x04},
		[]byte{0xAB, 0xCD},
		[]byte{0xAB, 0xFE},
	))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	msg, err := r.ReadMessage()

	require.NoError(t, err)
	assert.Len(t, msg, 4)
}

func TestReadMessage_ReadReturns_BytesAndEOF(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00, 0x00, 0x00},
		[]byte{0xAB},
	))

	_, err := r.ReadHeader()
	require.NoError(t, err)

	msg, err := r.ReadMessage()

	require.NoError(t, err)
	assert.Len(t, msg, 0)
}

func newTestReader(messages ...[]byte) *bytes.Buffer {
	var long []byte
	for _, msg := range messages {
		long = append(long, msg...)
	}
	return bytes.NewBuffer(long)
}
