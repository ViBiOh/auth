package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	flag.Parse()

	if len(os.Args) != 3 {
		fmt.Println("Usage is bcrypt [password] [cost]")
		os.Exit(1)
	}

	cost, err := strconv.ParseInt(os.Args[2], 10, 32)
	logger.FatalfOnErr(context.Background(), err, "parse cost")

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), int(cost))
	logger.FatalfOnErr(context.Background(), err, "generate password")

	fmt.Print(string(hash))
}
