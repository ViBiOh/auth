package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	flag.Parse()

	if len(os.Args) != 3 {
		fmt.Println("Usage is bcrypt [password] [cost]")
		os.Exit(1)
	}

	ctx := context.Background()

	cost, err := strconv.ParseInt(os.Args[2], 10, 32)
	logger.FatalfOnErr(ctx, err, "parse cost")

	rawPassword := []byte(os.Args[1])
	if len(rawPassword) > ident.MaxPasswordLength {
		logger.FatalfOnErr(ctx, err, "check password")
	}

	hash, err := bcrypt.GenerateFromPassword(rawPassword, int(cost))
	logger.FatalfOnErr(ctx, err, "generate password")

	fmt.Print(string(hash))
}
