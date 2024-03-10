package account

import (
	"crypto/sha1"
	"encoding"
	"encoding/base64"
	"errors"
	"strings"

	_ "github.com/alecthomas/kong"
)

// NOTE: Password should not be a part of this struct
type Account struct {
	Username string `arg:"" help:"username of the account"`
	Site     string `arg:"" help:"site associated to the account"`
}

func New(username, site string) Account {
	return Account{
		Username: username,
		Site:     site,
	}
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
