package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"golang.org/x/term"
)

func main() {
	outDir := flag.String("keystore", "./keystore", "directory for keystore files")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o700); err != nil {
		panic(err)
	}

	fmt.Print("Enter new keystore password: ")
	pw, err := term.ReadPassword(int(os.Stderr.Fd()))
	if err != nil {
		panic(err)
	}

	ks := keystore.NewKeyStore(*outDir, keystore.StandardScryptN, keystore.StandardScryptP)
	acct, err := ks.NewAccount(string(pw))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created account: %s\n", acct.Address.Hex())
	fmt.Printf("Keystore dir:   %s\n", *outDir)
	fmt.Println("NOTE: Back up the keystore file and remember the password.")
}
