package main

import (
	"crypto/sha1"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

type Account struct {
	Username string `help:"Username" short:"u" json:"user"`
	Password string `help:"Password" short:"p" json:"pass"`
	Site     string `help:"Site" short:"s" json:"site"`
}

const DB_PATH string = ".passman-db.json"

var (
	dbFile   *os.File
	accounts map[string]Account = map[string]Account{}
)

func init() {
	// get home path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dbPath := fmt.Sprintf("%s%s%s", homeDir, string(os.PathSeparator), DB_PATH)

	// open file
	dbFile, err = os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// read file
	data, err := os.ReadFile(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(data) == 0 {
		return
	}

	if err := json.Unmarshal(data, &accounts); err != nil {
		log.Fatal("error reading db(json): " + err.Error())
	}
}

type CLI struct {
	New Account `cmd:"" help:"creating new account"`

	Ls struct {
		Oneline bool `cmd:"" help:"Printing in a line"`
	} `cmd:"ls"`

	Get struct {
		Username string `cmd:"" short:"u"`
		Site     string `cmd:"" short:"s"`
	} `cmd:"get"`
}

func main() {
	defer dbFile.Close()

	cli := &CLI{}
	ctx := kong.Parse(cli)

	var err error
	switch ctx.Command() {
	case "ls":
		log.Println("LS")

	case "new":
		err = newAccount(&cli.New)
		if err != nil {
			err = fmt.Errorf("Failed to create new account:\n\t%v", err)
		}

	case "get":
		err = getAccount(cli)
		if err != nil {
			err = fmt.Errorf("Failed to create new account:\n\t%v", err)
		}

	default:
		log.Fatal("not implemented: " + ctx.Command())
	}

	if err != nil {
		fmt.Println(err)
	}
}

func getAccount(cli *CLI) error {
	id, err := getAccountID(cli.Get.Username, cli.Get.Site)
	if err != nil {
		return err
	}

	acc, exists := accounts[id]
	if !exists {
		return errors.New("Account not found")
	}

	log.Println(acc)

	return nil
}

func getAccountID(username string, site string) (string, error) {
	acc := Account{
		Username: username,
		Site:     site,
	}
	return acc.ID()
}

func newAccount(account *Account) error {
	// creating id
	id, err := account.ID()
	if err != nil {
		return fmt.Errorf("Error creating id: %v", err)
	}

	// check for duplication
	_, exists := accounts[id]
	if exists {
		return fmt.Errorf("Site exists: %s", account.Site)
	}

	// creating new account
	accounts[id] = *account
	b, err := json.MarshalIndent(accounts, "", "    ")
	if err != nil {
		return fmt.Errorf("Error marshalling bytes %v", err)
	}
	_, err = dbFile.Write(b)
	if err != nil {
		return fmt.Errorf("Failed to write: %v", err)

	}
	return nil
}

func (acc *Account) ID() (string, error) {
	b := sha1.New()
	parts := acc.Username + acc.Site
	_, err := b.Write([]byte(strings.ToLower(parts)))
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

	id := base64.RawStdEncoding.EncodeToString(idBytes)
	return id, nil
}
