package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/hn275/pass-man/internal/account"
	"github.com/hn275/pass-man/internal/crypto"
	"github.com/hn275/pass-man/internal/database"
	"github.com/mattn/go-sqlite3"
)

const DB_PATH string = ".passman-db.json"

type CLI struct {
	New account.Account `cmd:"" help:"creating new account."`
	Get struct{}        `cmd:"" help:"selecting and writing password to clipboard buffer."`
}

func main() {
	db := database.New()
	defer db.Close()

	cli := &CLI{}
	ctx := kong.Parse(cli)

	var err error
	switch ctx.Command() {
	case "new <username> <site>":
		err = handleNewAccount(&cli.New)
		if err != nil {
			err = fmt.Errorf("Failed to create new account:\n%v", err)
		}

	case "get":
		err = getAccount()
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

type Credential struct {
	ID   string `db:"id"`
	User string `db:"user"`
	Site string `db:"site"`
	Pass string `db:"pass"`
}

func (c *Credential) String() string {
	return fmt.Sprintf("%s @ %s", c.User, c.Site)
}

func getAccount() error {
	db := database.New()
	defer db.Close()

	// get all accounts
	var accounts []Credential
	rows, err := db.Queryx("SELECT id,user,site,pass FROM pass;")
	if err != nil {
		return err
	}

	for rows.Next() {
		var cred Credential
		if err := rows.StructScan(&cred); err != nil {
			return err
		}
		accounts = append(accounts, cred)
	}

	if len(accounts) == 0 {
		fmt.Println("You have no saved accounts.")
		os.Exit(0)
	}

	// get user input
	for i, acc := range accounts {
		fmt.Printf("%d %s\n", i+1, acc.String())
	}

	fmt.Println()
	var selectedAccount Credential
	for {
		fmt.Printf("Select an account (default=1): ")
		buf := make([]byte, 10)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			printErrExit(err.Error())
		}

		buf = buf[:n-1] // since the last char is a new line
		if len(buf) == 0 {
			selectedAccount = accounts[0]
			break
		}

		choice, err := strconv.Atoi(string(buf))
		if err != nil {
			fmt.Println("Invalid number, try again.")
			log.Println(err)
			continue
		}

		if choice > len(accounts) || choice <= 0 {
			fmt.Println("Invalid selection, try again.")
			continue
		}

		selectedAccount = accounts[choice-1]
		break
	}

	return nil
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
