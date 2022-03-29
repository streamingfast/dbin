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
			contentType: "eth",
			expect:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x03, 'e', 't', 'h'},
		},
		{
			name:        "content type too long",
			contentType: strings.Repeat("eth", 100000),
			expectError: "content type too long, expected maximum 65535 in length, found 300000 bytes",
		},
		{
			name:        "empty content type",
			contentType: "",
			expect:      []byte{'d', 'b', 'i', 'n', 0x01, 0x00, 0x00},
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
	err := w.WriteHeader("eth") /* such leet */
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
	err := w.WriteHeader("TES")
	require.NoError(t, err)

	err = w.WriteHeader("AGA") /* leet speech for again :) */

	require.Error(t, err)
	assert.Equal(t, "header already written", err.Error())
}

func TestReadWrite(t *testing.T) {
	msg1 := "hello world"
	msg2 := "my friend is great"

	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	require.NoError(t, w.WriteHeader("TES75"))
	require.NoError(t, w.WriteMessage([]byte(msg1)))
	require.NoError(t, w.WriteMessage([]byte(msg2)))
	require.NoError(t, w.Close())

	r := NewReader(bytes.NewReader(buf.Bytes()))
	contentType, err := r.ReadHeader()
	require.NoError(t, err)
	assert.Equal(t, "TES75", contentType)

	back1, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg1, string(back1))

	back2, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg2, string(back2))
}
