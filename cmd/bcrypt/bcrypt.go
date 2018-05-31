package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/bcrypt"
)

func getPromptPassword() string {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error while reading prompt: %v", err)
		os.Exit(1)
	}

	return result
}

func getArgsPassword() string {
	if len(os.Args) != 2 {
		fmt.Println(`Usage is bcrypt [password]`)
		os.Exit(1)
	}

	return os.Args[1]
}

func main() {
	prompt := flag.Bool(`prompt`, false, `Prompt for password rather than reading CLI args`)
	flag.Parse()

	var password string

	if *prompt {
		password = getPromptPassword()
	} else {
		password = getArgsPassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Printf(`Error while generating password: %v`, err)
	}
	fmt.Print(string(hash))
}
