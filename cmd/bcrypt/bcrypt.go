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

	bestCost, err := findBestCost(time.Second / 5)
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
	var duration time.Duration
	password := []byte("b6aa8c7d9931406946efe9ba2fadc1a6") // random string

	cost := bcrypt.MinCost
	for ; duration < maxDuration && cost < bcrypt.MaxCost; cost++ {
		start := time.Now()

		if _, err := bcrypt.GenerateFromPassword(password, cost); err != nil {
			return cost, fmt.Errorf("unable to generate password: %s", err)
		}

		duration = time.Since(start)
	}

	return cost, nil
}
