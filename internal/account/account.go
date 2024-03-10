package account

import (
	"crypto/sha1"
	"encoding"
	"encoding/base64"
	"errors"
	"strings"
)

type Account struct {
	Username string `help:"Username" short:"u" json:"user"`
	Site     string `help:"Site" short:"s" json:"site"`
	Password string `json:"pass"`
}

func (acc *Account) ID() (string, error) {
	b := sha1.New()
	id := acc.Username + acc.Site
	_, err := b.Write([]byte(strings.ToLower(id)))
	if err != nil {
		return "", err
	}

	binMarshaler, ok := b.(encoding.BinaryMarshaler)
	if !ok {
		return "", errors.New("BinaryMarshaler not implemented")
	}

	idBytes, err := binMarshaler.MarshalBinary()
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(idBytes), nil
}
