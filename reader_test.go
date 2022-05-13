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
		name          string
		input         []byte
		expectType    string
		expectVersion int32
		expectError   string
	}{
		{
			name:          "happy path",
			input:         []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
			expectType:    "ETH",
			expectVersion: 98,
		},
		{
			name:        "bad prefix",
			input:       []byte{'d', 'b', 'o', 'b', 0x00, 'E', 'T', 'H', '9', '8'},
			expectError: "magic string 'dbin' not found in header",
		},
		{
			name:        "bad library revision, or file format version",
			input:       []byte{'d', 'b', 'i', 'n', 0x01, 'E', 'T', 'H', '9', '8'},
			expectError: "invalid dbin file revision, expected 0, got 1",
		},
		{
			name:        "incomplete header",
			input:       []byte{'d', 'b', 'i', 'n', 0x01, 'E', 'T', 'H'},
			expectError: "unexpected EOF",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(test.input))

			contentType, version, err := r.ReadHeader()

			if test.expectError == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expectType, contentType)
				assert.Equal(t, test.expectVersion, version)
			} else {
				require.Error(t, err)
				assert.Equal(t, test.expectError, err.Error())
				assert.Equal(t, "", contentType)
				assert.Equal(t, int32(0), version)
			}
			assert.NoError(t, r.Close())
		})
	}
}

func TestDoubleReadHeader(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'}))
	r.ReadHeader()

	_, _, err := r.ReadHeader()

	assert.Error(t, err)
	assert.Equal(t, "header was read already", err.Error())
}

func TestReadMessageHappy(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{
		'd', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8',
		0x00, 0x00, 0x00, 0x01, // msg 1, length
		0x61,
	}))
	_, _, err := r.ReadHeader()
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
	_, _, err := r.ReadHeader()
	require.NoError(t, err)

	message, err := r.ReadMessage()
	assert.Nil(t, message)
	assert.Equal(t, io.EOF, err)
}

func TestReadMessage_IncompleteMesageLength(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00, 0x00},
	))

	_, _, err := r.ReadHeader()
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

	_, _, err := r.ReadHeader()
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

	_, _, err := r.ReadHeader()
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

	contentType, version, err := r.ReadHeader()
	require.NoError(t, err)

	assert.Equal(t, "ETH", contentType)
	assert.Equal(t, int32(98), version)
}

func TestReadMessage_MesageLength_MultiReadNeeded(t *testing.T) {
	r := NewReader(newTestReader(
		[]byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		[]byte{0x00, 0x00},
		[]byte{0x00, 0x00},
	))

	_, _, err := r.ReadHeader()
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

	_, _, err := r.ReadHeader()
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

	_, _, err := r.ReadHeader()
	require.NoError(t, err)

	msg, err := r.ReadMessage()

	require.NoError(t, err)
	assert.Len(t, msg, 0)
}

type testReader struct {
	messages [][]byte
	index    int
}

func newTestReader(messages ...[]byte) *testReader {
	return &testReader{
		messages: messages,
	}
}

func (r *testReader) Read(p []byte) (n int, err error) {
	if r.index >= len(r.messages) {
		return 0, io.EOF
	}

	message := r.messages[r.index]
	r.index++

	count := copy(p, message)

	if r.index == len(r.messages) {
		return count, io.EOF
	}

	return count, nil
}
