package crypto_test

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/hn275/pass-man/internal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	pt := []byte("Hello world!")
	ad := make([]byte, 100)

	key := make([]byte, crypto.KeySize)
	_, err := io.ReadFull(rand.Reader, key)
	assert.Nil(t, err)

	cipher := crypto.New(key)

	ciphertext, err := cipher.Encrypt(pt, ad)
	assert.Nil(t, err)
	assert.NotEqual(t, pt, ciphertext)

	plaintext, err := cipher.Decrypt(ciphertext, ad)
	assert.Nil(t, err)
	assert.Equal(t, plaintext, pt)
}
