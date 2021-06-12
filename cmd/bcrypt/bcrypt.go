package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	flag.Parse()

	if len(os.Args) != 2 {
		fmt.Println("Usage is bcrypt [password]")
		os.Exit(1)
	}

	bestCost, err := findBestCost(time.Second / 4)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Best bcrypt cost is %d", bestCost)

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bestCost)
	if err != nil {
		fmt.Printf("generate password: %v", err)
	}
	fmt.Print(string(hash))
}

func findBestCost(maxDuration time.Duration) (int, error) {
	password := []byte("b6aa8c7d9931406946efe9ba2fadc1a6") // random string

	for i := bcrypt.MinCost + 1; i <= bcrypt.MaxCost; i++ {
		hashedPassword, err := bcrypt.GenerateFromPassword(password, i)
		if err != nil {
			return i, fmt.Errorf("unable to generate password: %s", err)
		}

		start := time.Now()
		_ = bcrypt.CompareHashAndPassword(hashedPassword, password)
		duration := time.Since(start)

		if duration > maxDuration {
			return i - 1, nil
		}
	}

	return bcrypt.MaxCost, nil
}
