package basic

import (
	"context"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/tools"
	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	*model.User
	password []byte
}

// Config of package
type Config struct {
	users *string
}

// App of package
type App struct {
	users map[string]*basicUser
	db    *sql.DB
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		users: fs.String(tools.ToCamel(fmt.Sprintf("%sUsers", prefix)), "", "[basic] Users in the form `id:username:password,id2:username2:password2`"),
	}
}

// New creates new App from Config
func New(config Config, db *sql.DB) (ident.Auth, error) {
	users, err := loadUsers(*config.users)
	if err != nil {
		return nil, err
	}

	return &App{
		users: users,
		db:    db,
	}, nil
}

func loadUsers(authUsers string) (map[string]*basicUser, error) {
	if authUsers == "" {
		return nil, nil
	}

	users := make(map[string]*basicUser)

	for _, authUser := range strings.Split(authUsers, ",") {
		parts := strings.Split(authUser, ":")
		if len(parts) != 3 {
			return nil, errors.New("invalid format of user for %s", authUser)
		}

		user := basicUser{&model.User{ID: parts[0], Username: strings.ToLower(parts[1])}, []byte(parts[2])}
		users[strings.ToLower(user.Username)] = &user
	}

	return users, nil
}

// GetName returns Authorization header prefix
func (App) GetName() string {
	return "Basic"
}

// GetUser returns User associated to header
func (a App) GetUser(ctx context.Context, header string) (*model.User, error) {
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, ":")
	if sepIndex < 0 {
		return nil, errors.New("invalid format for basic auth")
	}

	username := strings.ToLower(dataStr[:sepIndex])
	password := dataStr[sepIndex+1:]

	if a.users == nil && a.db == nil {
		return nil, errors.New("no basic source provided")
	}

	var user *basicUser
	if a.users != nil {
		user = a.users[username]
	} else {
		user = a.dbLoginUser(username)
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user.User, nil
}

// Redirect redirects user to login endpoint
func (App) Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login/basic", http.StatusFound)
}

// Login exchange state to token
func (a App) Login(r *http.Request) (string, error) {
	authContent := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), a.GetName()))

	if _, err := a.GetUser(r.Context(), authContent); err != nil {
		return "", err
	}
	return authContent, nil
}

// OnLoginError handle action when login fails
func (App) OnLoginError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")
	httperror.Unauthorized(w, err)
}
