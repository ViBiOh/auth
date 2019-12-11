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
	Unmarsall(data []byte) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
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
func (a app) Unmarsall(data []byte) (interface{}, error) {
	var user model.User

	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return user, nil
}

// List Users
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	if err := a.checkAdminOrSelfContext(ctx, 0); err != nil {
		return nil, 0, err
	}

	list, total, err := a.list(page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %w", err)
	}

	itemsList := make([]interface{}, len(list))
	for index, item := range list {
		itemsList[index] = item
	}

	return itemsList, total, nil
}

// Get User
func (a app) Get(ctx context.Context, ID uint64) (interface{}, error) {
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
func (a app) Create(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	id, err := a.create(user, nil)
	if err != nil {
		return model.NoneUser, fmt.Errorf("unable to create: %w", err)
	}

	user.ID = id

	return user, nil
}

// Update User
func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	if err := a.checkAdminOrSelfContext(ctx, user.ID); err != nil {
		return nil, err
	}

	if err := a.update(user, nil); err != nil {
		return nil, fmt.Errorf("unable to update: %w", err)
	}

	return o, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	user := o.(model.User)

	if err := a.checkAdminOrSelfContext(ctx, user.ID); err != nil {
		return err
	}

	if err := a.delete(user, nil); err != nil {
		err = fmt.Errorf("unable to delete: %w", err)
	}

	return
}

func (a app) Check(ctx context.Context, old, new interface{}) []error {
	output := make([]error, 0)

	if new == nil && a.auth.IsAuthorized(handler.UserFromContext(ctx), "admin") {
		output = append(output, errors.New("you must be an admin to delete user"))
	}

	if new == nil {
		return output
	}

	newUser := new.(model.User)

	if strings.TrimSpace(newUser.Login) == "" {
		output = append(output, fmt.Errorf("name is required: %w", crud.ErrInvalid))
	}

	if old == nil && new != nil && strings.TrimSpace(newUser.Password) == "" {
		output = append(output, fmt.Errorf("password is required: %w", crud.ErrInvalid))
	}

	return output
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
