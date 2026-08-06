package postgres

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKey() []byte {
	return bytes.Repeat([]byte("k"), 32)
}

func TestEncryptDecryptField_RoundTrip(t *testing.T) {
	key := testKey()
	ciphertext, err := encryptField(key, "https://push.example.com/abc123")
	require.NoError(t, err)
	assert.NotEqual(t, "https://push.example.com/abc123", ciphertext, "ciphertext must not equal plaintext")

	plaintext, err := decryptField(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "https://push.example.com/abc123", plaintext)
}

func TestEncryptField_NonDeterministic(t *testing.T) {
	key := testKey()
	a, err := encryptField(key, "same input")
	require.NoError(t, err)
	b, err := encryptField(key, "same input")
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "AES-GCM uses a random nonce, so repeated encryption of the same plaintext must differ")
}

func TestDecryptField_WrongKeyFails(t *testing.T) {
	ciphertext, err := encryptField(testKey(), "secret value")
	require.NoError(t, err)

	wrongKey := bytes.Repeat([]byte("x"), 32)
	_, err = decryptField(wrongKey, ciphertext)
	assert.Error(t, err)
}

func TestHashEndpoint_Deterministic(t *testing.T) {
	a := hashEndpoint("https://push.example.com/abc123")
	b := hashEndpoint("https://push.example.com/abc123")
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, hashEndpoint("https://push.example.com/different"))
}
