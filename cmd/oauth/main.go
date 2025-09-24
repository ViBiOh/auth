package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const jwtExpiration = time.Hour * 24 * 5

var (
	signMethod       = jwt.SigningMethodHS256
	signValidMethods = []string{signMethod.Alg()}
	cookieMaxAge     = int(jwtExpiration.Seconds())
)

var hmacSecret = []byte("strong_secret")

type User struct {
	Login string
	ID    int
}

type AuthClaims struct {
	Login string        `json:"login"`
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
}

func main() {
	config := oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://127.0.0.1:1080/auth/github/callback",
		Scopes:       nil,
	}

	verifier := oauth2.GenerateVerifier()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier)), http.StatusFound)
	})

	http.HandleFunc("/auth/github/check", func(w http.ResponseWriter, r *http.Request) {
		auth, err := r.Cookie("auth")
		if err != nil {
			http.Error(w, "get auth cookie: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var claim AuthClaims

		if _, err = jwt.ParseWithClaims(auth.Value, &claim, func(token *jwt.Token) (any, error) { return hmacSecret, nil }, jwt.WithValidMethods(signValidMethods)); err != nil {
			http.Error(w, "parse JWT: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintf(w, "token=%s, name=%s", claim.Token.AccessToken, claim.Login)
	})

	http.HandleFunc("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		oauth2Token, err := config.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, "exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.Client(r.Context(), oauth2Token)
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			http.Error(w, "get /user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := httpjson.Read[User](resp)
		if err != nil {
			http.Error(w, "read /user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		token := jwt.NewWithClaims(signMethod, newClaim(oauth2Token, user))

		tokenString, err := token.SignedString(hmacSecret)
		if err != nil {
			http.Error(w, "sign JWT: "+err.Error(), http.StatusInternalServerError)
			return
		}

		setCallbackCookie(w, r, "auth", tokenString)

		_, _ = fmt.Fprintf(w, "Visit http://127.0.0.1:1080/auth/github/check")
	})

	log.Printf("listening on http://%s/", "127.0.0.1:1080")
	log.Fatal(http.ListenAndServe("127.0.0.1:1080", nil))
}

func newClaim(token *oauth2.Token, user User) AuthClaims {
	now := time.Now()

	return AuthClaims{
		Login: user.Login,
		Token: token,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth",
			Subject:   user.Login,
			ID:        strconv.Itoa(user.ID),
		},
	}
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   cookieMaxAge,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
		HttpOnly: true,
	})
}
