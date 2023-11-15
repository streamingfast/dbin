package dbin

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteHeader(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expect      []byte
		expectError string
	}{
		{
			name:        "happy path",
			contentType: "type.googleapis.com/sf.ethereum.type.v2.Block",
			expect:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x2D, 't', 'y', 'p', 'e', '.', 'g', 'o', 'o', 'g', 'l', 'e', 'a', 'p', 'i', 's', '.', 'c', 'o', 'm', '/', 's', 'f', '.', 'e', 't', 'h', 'e', 'r', 'e', 'u', 'm', '.', 't', 'y', 'p', 'e', '.', 'v', '2', '.', 'B', 'l', 'o', 'c', 'k'},
		},
		{
			name:        "content type too long",
			contentType: strings.Repeat("E", maxContentTypeLength+1),
			expectError: "contentType length should not exceed 65535 characters, was 65536",
		},
		{
			name:        "content type too short",
			contentType: "",
			expectError: "contentType should contain at-least one character",
		},
		{
			name:        "that's 3 chars UTF-8 dude",
			contentType: "ÉT",
			expect:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x03, 0xc3, 0x89, 'T'},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			err := w.WriteHeader(test.contentType)

			if test.expectError == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expect, buf.Bytes())
			} else {
				require.Error(t, err)
				assert.Equal(t, test.expectError, err.Error())
			}

			require.NoError(t, w.Close())
		})
	}
}

func TestWriteMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.WriteHeader("eth")
	require.NoError(t, err)

	err = w.WriteMessage([]byte("pouille"))
	require.NoError(t, err)
	err = w.WriteMessage([]byte("mouille"))
	require.NoError(t, err)

	assert.Equal(t, []byte{
		'd', 'b', 'i', 'n', 0x01, 0x00, 0x03, 'e', 't', 'h',
		0x00, 0x00, 0x00, 0x07,
		'p', 'o', 'u', 'i', 'l', 'l', 'e',
		0x00, 0x00, 0x00, 0x07,
		'm', 'o', 'u', 'i', 'l', 'l', 'e',
	}, buf.Bytes())
	assert.NoError(t, w.Close())
}

func TestWriteHeaderDouble(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.WriteHeader("type.googleapis.com/sf.antelope.type.v1.Block")
	require.NoError(t, err)

	err = w.WriteHeader("type.googleapis.com/sf.antelope.type.v1.Block")

	require.Error(t, err)
	assert.Equal(t, "header already written", err.Error())
}

func TestReadWrite(t *testing.T) {
	msg1 := "hello world"
	msg2 := "my friend is great"

	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	require.NoError(t, w.WriteHeader("type.googleapis.com/sf.solana.type.v1.Block"))
	require.NoError(t, w.WriteMessage([]byte(msg1)))
	require.NoError(t, w.WriteMessage([]byte(msg2)))
	require.NoError(t, w.Close())

	r := NewReader(bytes.NewReader(buf.Bytes()))
	contentType, err := r.ReadHeader()
	require.NoError(t, err)
	assert.Equal(t, "type.googleapis.com/sf.solana.type.v1.Block", contentType)

	back1, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg1, string(back1))

	back2, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg2, string(back2))
}
