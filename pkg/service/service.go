package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/handler"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var _ crud.Service = &app{}

// App of package
type App interface {
	Unmarsall(data []byte) (crud.Item, error)
	Check(old, new crud.Item) []error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]crud.Item, uint, error)
	Get(ctx context.Context, ID uint64) (crud.Item, error)
	Create(ctx context.Context, o crud.Item) (crud.Item, uint64, error)
	Update(ctx context.Context, o crud.Item) (crud.Item, error)
	Delete(ctx context.Context, o crud.Item) error
}

type app struct {
	db   *sql.DB
	auth auth.Provider
}

// New creates new App from Config
func New(db *sql.DB, auth auth.Provider) App {
	return &app{
		db:   db,
		auth: auth,
	}
}

// Unmarsall User
func (a app) Unmarsall(data []byte) (crud.Item, error) {
	var user model.User

	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// List Users
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]crud.Item, uint, error) {
	if err := a.checkAdminOrSelfContext(ctx, 0); err != nil {
		return nil, 0, err
	}

	list, total, err := a.list(page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %w", err)
	}

	itemsList := make([]crud.Item, len(list))
	for index, item := range list {
		itemsList[index] = &item
	}

	return itemsList, total, nil
}

// Get User
func (a app) Get(ctx context.Context, ID uint64) (crud.Item, error) {
	if err := a.checkAdminOrSelfContext(ctx, ID); err != nil {
		return nil, err
	}

	item, err := a.get(ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, crud.ErrNotFound
		}
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	return &item, nil
}

// Create User
func (a app) Create(ctx context.Context, o crud.Item) (crud.Item, uint64, error) {
	id, err := a.create(*o.(*model.User), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create: %w", err)
	}

	return o, id, nil
}

// Update User
func (a app) Update(ctx context.Context, o crud.Item) (crud.Item, error) {
	user := o.(*model.User)

	if err := a.checkAdminOrSelfContext(ctx, user.ID); err != nil {
		return nil, err
	}

	if err := a.update(*user, nil); err != nil {
		return nil, fmt.Errorf("unable to update: %w", err)
	}

	return o, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o crud.Item) (err error) {
	user := o.(*model.User)

	if err := a.checkAdminOrSelfContext(ctx, user.ID); err != nil {
		return err
	}

	if err := a.delete(*user, nil); err != nil {
		err = fmt.Errorf("unable to delete: %w", err)
	}

	return
}

func (a app) Check(old, new crud.Item) []error {
	if new == nil {
		return nil
	}

	user := old.(*model.User)
	errors := make([]error, 0)

	if strings.TrimSpace(user.Login) == "" {
		errors = append(errors, fmt.Errorf("name is required: %w", crud.ErrInvalid))
	}

	return errors
}

func (a app) checkAdminOrSelfContext(ctx context.Context, id uint64) error {
	user := handler.UserFromContext(ctx)
	if user == model.NoneUser {
		return crud.ErrUnauthorized
	}

	if id != 0 && user.ID == id {
		return nil
	}

	if !a.auth.IsAuthorized(user, "admin") {
		return crud.ErrForbidden
	}

	return nil
}
