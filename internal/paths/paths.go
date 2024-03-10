package paths

import (
	"fmt"
	"io/fs"
	"log"
	"os"
)

var homeDir string

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("user home directory not found: %v\n", err)
	}

	// check for pass-man home directory existence
	homeDir = homeDir + string(os.PathSeparator) + ".pass-man"
	_, err = os.Stat(homeDir)
	switch err.(type) {
	case nil:
		break

	case *fs.PathError:
		fmt.Printf("Creating data directory\n")
		err = os.Mkdir(homeDir, 0777)
		if err == nil {
			fmt.Printf("Path created: %s\n", homeDir)
		}

	default:
		fmt.Printf("Failed to create data directory: %v\n", err)
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("failed to create data directory: %v\n", err)
	}
}

func MakePath(fileName string) string {
	return homeDir + string(os.PathSeparator) + fileName
}
