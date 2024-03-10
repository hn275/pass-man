package account

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
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
	b := sha256.New()
	id := acc.Username + acc.Site
	_, err := b.Write([]byte(strings.ToLower(id)))
	if err != nil {
		return "", err
	}

	idBytes := b.Sum(nil)
	return hex.EncodeToString(idBytes), nil
}
