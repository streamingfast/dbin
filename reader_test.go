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
			name:       "happy path",
			input:      []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
			expectType: "ETH98",
		},
		{
			name:        "bad prefix",
			input:       []byte{'d', 'b', 'o', 'b', 0x00, 'E', 'T', 'H', '9', '8'},
			expectError: "magic string 'dbin' not found in header",
		},
		{
			name:        "proto type with file version 1, too short",
			input:       []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x19, 'p', 'r', 'o', 't'},
			expectError: "reading content type of length 25: unexpected EOF",
		},
		{
			name:       "proto type with file version 1, empty",
			input:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x00},
			expectType: "",
		},
		{
			name:       "proto type with file version 1, happy",
			input:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x05, 'p', 'r', 'o', 't', 'o', 0x00},
			expectType: "proto",
		},
		{
			name:        "bad library revision, or file format version",
			input:       []byte{'d', 'b', 'i', 'n', 0x02, 'w', 'h', 'a', 't', 'e', 'v', 'e', 'r'},
			expectError: "invalid dbin file version, expected 0 or 1, got 2",
		},
		{
			name:        "incomplete header",
			input:       []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H'},
			expectError: "reading content type of file v0: unexpected EOF",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(test.input))

			contentType, err := r.ReadHeader()

			if test.expectError == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expectType, contentType)
			} else {
				require.Error(t, err)
				assert.Equal(t, test.expectError, err.Error())
				assert.Equal(t, "", contentType)
			}
			assert.NoError(t, r.Close())
		})
	}
}

func TestDoubleReadHeader(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'}))
	_, err := r.ReadHeader()
	require.NoError(t, err)

	_, err = r.ReadHeader()

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

	typ, err := r.ReadHeader()
	assert.Equal(t, typ, "ETH98")
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

func TestReadMessage_Header_MultiReadNeeded(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00},
		[]byte{'E', 'T', 'H', '9', '8'},
	))

	contentType, err := r.ReadHeader()
	require.NoError(t, err)

	assert.Equal(t, "ETH98", contentType)
}

func TestReadMessage_MessageLength_MultiReadNeeded(t *testing.T) {
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

// type testReader struct {
// 	messages [][]byte
// 	index    int
// }

func newTestReader(messages ...[]byte) *bytes.Buffer {
	var long []byte
	for _, msg := range messages {
		long = append(long, msg...)
	}
	return bytes.NewBuffer(long)
}

// func (r *testReader) Read(p []byte) (n int, err error) {
// 	if r.index >= len(r.messages) {
// 		return 0, io.EOF
// 	}

// 	message := r.messages[r.index]
// 	r.index++

// 	count := copy(p, message)

// 	if r.index == len(r.messages) {
// 		return count, io.EOF
// 	}

// 	return count, nil
// }
