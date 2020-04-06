package dbin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteHeader(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		version     int
		expect      []byte
		expectError string
	}{
		{
			name:        "happy path",
			contentType: "ETH",
			version:     98,
			expect:      []byte{'d', 'b', 'i', 'n', 0x00, 'E', 'T', 'H', '9', '8'},
		},
		{
			name:        "content type too long",
			contentType: "ETHereuuummmz",
			version:     98,
			expectError: "contentType should be 3 characters, was 13 [69 84 72 101 114 101 117 117 117 109 109 109 122]",
		},
		{
			name:        "content type too short",
			contentType: "À",
			version:     98,
			expectError: "contentType should be 3 characters, was 2 [195 128]",
		},
		{
			name:        "that's 3 chars UTF-8 dude",
			contentType: "ÉT",
			version:     0,
			expect:      []byte{'d', 'b', 'i', 'n', 0x00, 0xc3, 0x89, 'T', '0', '0'},
		},
		{
			name:        "version out of range low",
			contentType: "ETH",
			version:     -1,
			expectError: "version should be between 0 and 99, was -1",
		},
		{
			name:        "version out of range high",
			contentType: "ETH",
			version:     100,
			expectError: "version should be between 0 and 99, was 100",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			err := w.WriteHeader(test.contentType, test.version)

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
	err := w.WriteHeader("TES", 75) /* such leet */
	require.NoError(t, err)

	err = w.WriteMessage([]byte("pouille"))
	require.NoError(t, err)
	err = w.WriteMessage([]byte("mouille"))
	require.NoError(t, err)

	assert.Equal(t, []byte{
		'd', 'b', 'i', 'n', 0x00, 'T', 'E', 'S', '7', '5',
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
	err := w.WriteHeader("TES", 0)
	require.NoError(t, err)

	err = w.WriteHeader("AGA", 14) /* leet speech for again :) */

	require.Error(t, err)
	assert.Equal(t, "header already written", err.Error())
}

func TestReadWrite(t *testing.T) {
	msg1 := "hello world"
	msg2 := "my friend is great"

	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	require.NoError(t, w.WriteHeader("TES", 75))
	require.NoError(t, w.WriteMessage([]byte(msg1)))
	require.NoError(t, w.WriteMessage([]byte(msg2)))
	require.NoError(t, w.Close())

	r := NewReader(bytes.NewReader(buf.Bytes()))
	contentType, version, err := r.ReadHeader()
	require.NoError(t, err)
	assert.Equal(t, "TES", contentType)
	assert.Equal(t, int32(75), version)

	back1, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg1, string(back1))

	back2, err := r.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg2, string(back2))
}
