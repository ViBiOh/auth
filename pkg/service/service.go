package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var _ crud.Service = &app{}

// App of package
type App interface {
	Unmarshal(data []byte, contentType string) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []crud.Error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error

	CheckRights(ctx context.Context, id uint64) error
}

type app struct {
	store store.UserStorage
	auth  auth.Provider
}

// New creates new App from Config
func New(store store.UserStorage, auth auth.Provider) App {
	return &app{
		store: store,
		auth:  auth,
	}
}

// Unmarshal User
func (a app) Unmarshal(data []byte, contentType string) (interface{}, error) {
	var user model.User

	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return user, nil
}

// List Users
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	if err := a.CheckRights(ctx, 0); err != nil {
		return nil, 0, err
	}

	list, total, err := a.store.List(ctx, page, pageSize, sortKey, sortAsc)
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
	if err := a.CheckRights(ctx, ID); err != nil {
		return nil, err
	}

	item, err := a.store.Get(ctx, ID)
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

	id, err := a.store.Create(ctx, user)
	if err != nil {
		return model.NoneUser, fmt.Errorf("unable to create: %w", err)
	}

	user.ID = id

	return user, nil
}

// Update User
func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	if err := a.CheckRights(ctx, user.ID); err != nil {
		return nil, err
	}

	if err := a.store.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("unable to update: %w", err)
	}

	return o, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	user := o.(model.User)

	if err := a.CheckRights(ctx, user.ID); err != nil {
		return err
	}

	if err := a.store.Delete(ctx, user); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return
}

func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	output := make([]crud.Error, 0)

	if new == nil && a.auth.IsAuthorized(ctx, model.ReadUser(ctx), "admin") {
		output = append(output, crud.NewError("profile", "you must be an admin to delete user"))
	}

	if new == nil {
		return output
	}

	newUser := new.(model.User)

	if strings.TrimSpace(newUser.Login) == "" {
		output = append(output, crud.NewError("name", "name is required"))
	}

	if old == nil && new != nil && strings.TrimSpace(newUser.Password) == "" {
		output = append(output, crud.NewError("password", "password is required"))
	}

	return output
}

func (a app) CheckRights(ctx context.Context, id uint64) error {
	user := model.ReadUser(ctx)
	if user == model.NoneUser {
		return crud.ErrUnauthorized
	}

	if id != 0 && user.ID == id || a.auth.IsAuthorized(ctx, user, "admin") {
		return nil
	}

	return crud.ErrForbidden
}
