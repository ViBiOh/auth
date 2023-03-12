package main

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/bcrypt"
)

func main() {
	fmt.Println(bcrypt.FindBestCost(time.Second / 4))
}
