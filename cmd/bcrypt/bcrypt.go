package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	flag.Parse()

	if len(os.Args) != 2 {
		fmt.Println("Usage is bcrypt [password]")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 12)
	if err != nil {
		fmt.Printf("generate password: %v", err)
	}
	fmt.Print(string(hash))
}
