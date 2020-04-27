package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
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
	err := json.Unmarshal(data, &user)
	return user, err
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
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneUser {
		return nil, crud.ErrNotFound
	}

	return item, nil
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

	if err := a.store.Update(ctx, user); err != nil {
		return user, fmt.Errorf("unable to update: %w", err)
	}

	return user, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o interface{}) error {
	user := o.(model.User)

	if err := a.store.Delete(ctx, user); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return nil
}

func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	output := make([]crud.Error, 0)

	user := model.ReadUser(ctx)
	if old != nil && user == model.NoneUser {
		output = append(output, crud.NewError("context", "you must be logged in for interacting"))
	}

	if new == nil && !a.auth.IsAuthorized(ctx, user, "admin") {
		output = append(output, crud.NewError("context", "you must be an admin for deleting"))
	}

	if new == nil {
		return output
	}

	newUser := new.(model.User)

	if old != nil && new != nil && !(user.ID == newUser.ID || a.auth.IsAuthorized(ctx, user, "admin")) {
		output = append(output, crud.NewError("context", "you're not authorized to interact with other user"))
	}

	if len(strings.TrimSpace(newUser.Login)) == 0 {
		output = append(output, crud.NewError("login", "login is required"))
	}

	if old == nil && new != nil && len(strings.TrimSpace(newUser.Password)) == 0 {
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

	logger.Info("unauthorized access for login=%s", user.Login)

	return crud.ErrForbidden
}
