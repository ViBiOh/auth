package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	var cookieValue string
	fmt.Println("Cookie value:")
	if _, err := fmt.Scanf("%s", &cookieValue); err != nil {
		log.Fatal(err)
	}

	var hmacSecret string
	fmt.Println("HMAC secret:")
	if _, err := fmt.Scanf("%s", &hmacSecret); err != nil {
		log.Fatal(err)
	}

	var claim cookie.Claim[model.OAuthClaim]
	if _, err := jwt.ParseWithClaims(cookieValue, &claim, func(t *jwt.Token) (any, error) { return []byte(hmacSecret), nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})); err != nil {
		log.Fatal(err)
	}

	payload, err := json.Marshal(claim)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", payload)
}
