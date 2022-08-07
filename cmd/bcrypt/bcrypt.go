package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	flag.Parse()

	if len(os.Args) != 3 {
		fmt.Println("Usage is bcrypt [password] [cost]")
		os.Exit(1)
	}

	cost, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Printf("parse cost: %s", err)
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), int(cost))
	if err != nil {
		fmt.Printf("generate password: %s", err)
		os.Exit(1)
	}
	fmt.Print(string(hash))
}
