package main

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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

func main() {
	fmt.Println(findBestCost(time.Second / 4))
}
