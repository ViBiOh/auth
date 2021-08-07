package db

import (
	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/httputils/v4/pkg/db"
)

// App of package
type App struct {
	db db.App
}

var (
	_ auth.Provider  = App{}
	_ auth.Storage   = App{}
	_ basic.Provider = App{}
)

// New creates new App from dependencies
func New(db db.App) App {
	return App{
		db: db,
	}
}
