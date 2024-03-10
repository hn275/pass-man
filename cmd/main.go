package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/hn275/pass-man/internal/account"
	"github.com/hn275/pass-man/internal/crypto"
	"github.com/hn275/pass-man/internal/database"
	"github.com/mattn/go-sqlite3"
)

const DB_PATH string = ".passman-db.json"

var (
	dbFile   *os.File
	accounts = make(map[string]account.Account)
)

type CLI struct {
	New account.Account `cmd:"new" help:"creating new account"`

	Ls struct {
		Oneline bool `cmd:"" help:"Printing in a line"`
	} `cmd:"ls"`

	Get account.Account `cmd:"get"`
}

func main() {
	defer dbFile.Close()

	cli := &CLI{}
	ctx := kong.Parse(cli)

	var err error
	switch ctx.Command() {
	case "ls":
		err = handleLS(cli)
		if err != nil {
			err = fmt.Errorf("Failed to list accounts:\n%v", err)
		}

	case "new":
		err = handleNewAccount(&cli.New)
		if err != nil {
			err = fmt.Errorf("Failed to create new account:\n%v", err)
		}

	case "get":
		err = getAccount(cli)
		if err != nil {
			err = fmt.Errorf("Failed to get an account:\n%v", err)
		}

	default:
		log.Fatal("not implemented: " + ctx.Command())
	}

	if err != nil {
		fmt.Println(err)
	}
}

func handleLS(cli *CLI) error {
	var err error
	_ = database.New()
	return err
}

func handleNewAccount(account *account.Account) error {
	if len(account.Username) == 0 || len(account.Site) == 0 {
		return errors.New("Invalid account detail")
	}

	fmt.Println("Enter account password:")
	pass, err := crypto.ReadSecretStdin()
	if err != nil {
		return err
	}

	// creating id
	id, err := account.ID()
	if err != nil {
		return fmt.Errorf("Error creating id: %v", err)
	}

	// scramble password
	ad, err := getMasterKey()
	if err != nil {
		printErrExit("error reading masterkey: %v", err)
	}

	cipher := crypto.New(crypto.SecretKey())
	ciphertext, err := cipher.Encrypt([]byte(pass), ad)
	if err != nil {
		printErrExit("error ciphering password: %v", err)
	}
	scrambledPass := hex.EncodeToString(ciphertext)

	// creating new account
	db := database.New()
	q := `INSERT INTO pass (id, user, pass, site) VALUES (?, ?, ?, ?);`
	_, err = db.Exec(q, id, account.Username, scrambledPass, account.Site)
	if err == nil {
		return nil
	}

	sqlite3err, ok := err.(sqlite3.Error)
	if ok && sqlite3err.Code == sqlite3.ErrConstraint {
		return errors.New("Account exists in database.")
	}
	return err
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
	acc := account.Account{
		Username: username,
		Site:     site,
	}
	return acc.ID()
}

func printErrExit(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
	os.Exit(1)
}

func getMasterKey() ([]byte, error) {
	fmt.Println("Enter master key:")
	key, err := crypto.ReadSecretStdin()
	if err != nil {
		return nil, err
	}
	return []byte(key), nil
}
