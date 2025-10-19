package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ViBiOh/auth/v3/pkg/argon"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	flag.Parse()

	if len(os.Args) != 2 {
		fmt.Println("Usage is argon [password]")
		os.Exit(1)
	}

	ctx := context.Background()

	encodedHash, err := argon.GenerateFromPassword(os.Args[1])
	logger.FatalfOnErr(ctx, err, "generate")

	fmt.Println(encodedHash)
}
