package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/term"
)

const (
	KeySize int = 32
)

type CipherBlock struct {
	key []byte
}

func New(key []byte) *CipherBlock {
	return &CipherBlock{key}
}

func (b *CipherBlock) Encrypt(plaintext, additionalData []byte) ([]byte, error) {
	var err error

	block, err := chacha20poly1305.NewX(b.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	buf := make([]byte, 0, block.NonceSize()+block.Overhead()+len(plaintext))
	buf = append(buf, nonce...)

	return block.Seal(buf, nonce, plaintext, additionalData), nil
}

func (b *CipherBlock) Decrypt(ciphertext, additionalData []byte) ([]byte, error) {
	block, err := chacha20poly1305.NewX(b.key)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:block.NonceSize()]

	ptSize := len(ciphertext) - block.NonceSize() - block.Overhead()
	if ptSize < 0 {
		return nil, errors.New("invalid ciphertext")
	}
	pt := make([]byte, 0, ptSize)

	encryptedText := ciphertext[block.NonceSize():]

	return block.Open(pt, nonce, encryptedText, additionalData)
}

func ReadSecretStdin(prompt string) (string, error) {
	fmt.Println(prompt)

	// Disable echoing input to the terminal
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	defer func() {
		err := term.Restore(int(os.Stdin.Fd()), oldState)
		if err != nil {
			log.Fatal("Error restoring terminal: " + err.Error())
		}
	}()

	// Read input without echoing to the screen
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(password), nil
}
