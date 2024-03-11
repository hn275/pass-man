package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hn275/pass-man/internal/paths"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/term"
)

const (
	KeySize int = 32
)

var (
	key = make([]byte, KeySize)
)

type CipherBlock struct {
	key []byte
}

func init() {
	keyFile := paths.MakePath("key")

	f, err := os.OpenFile(keyFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening key file: %v\n", err)
	}
	defer f.Close()

	buf := make([]byte, KeySize*2) // since the key is stored as hex
	_, err = f.Read(buf)
	switch err {
	case nil:
		key, err = hex.DecodeString(string(buf))
		if err != nil {
			fmt.Printf("error decoding key: %v\n", err)
			os.Exit(1)
		}
		if len(key) != KeySize {
			fmt.Println("unexpected key")
			os.Exit(1)
		}

	case io.EOF:
		fmt.Print("No key found, creating...")
		n, err := rand.Read(key)
		if err != nil {
			fmt.Printf("\nerror creating key: %v\n", err)
			os.Exit(1)
		}
		if n != KeySize {
			fmt.Printf("\nerror creating key: read %d/%d bytes\n", n, KeySize)
			os.Exit(1)
		}

		encoded := hex.EncodeToString(key)
		if _, err := f.Write([]byte(encoded)); err != nil {
			fmt.Printf("\n error writing new key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\tDone\n")

	default:
		fmt.Printf("error reading key: %v\n", err)
		os.Exit(1)
	}
}

func New(key []byte) *CipherBlock {
	return &CipherBlock{key}
}

func SecretKey() []byte {
	b := make([]byte, KeySize)
	n := copy(b, key)
	if n != KeySize {
		panic("error copying key")
	}
	return b
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

func ReadSecretStdin() (string, error) {
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
