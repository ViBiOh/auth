package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
)

// App of package
type App interface {
	Check(ctx context.Context, old, new model.User) error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]model.User, uint, error)
	Get(ctx context.Context, ID uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (model.User, error)
	Update(ctx context.Context, o model.User) (model.User, error)
	Delete(ctx context.Context, o model.User) error
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

// List Users
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, _ map[string][]string) ([]model.User, uint, error) {
	if err := a.CheckRights(ctx, 0); err != nil {
		return nil, 0, err
	}

	list, total, err := a.store.List(ctx, page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %w", err)
	}

	itemsList := make([]model.User, len(list))
	copy(itemsList, list)

	return itemsList, total, nil
}

// Get User
func (a app) Get(ctx context.Context, ID uint64) (model.User, error) {
	if err := a.CheckRights(ctx, ID); err != nil {
		return model.NoneUser, err
	}

	item, err := a.store.Get(ctx, ID)
	if err != nil {
		return model.NoneUser, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneUser {
		return model.NoneUser, httpModel.WrapNotFound(errors.New("user not found"))
	}

	return item, nil
}

// Create User
func (a app) Create(ctx context.Context, user model.User) (model.User, error) {
	id, err := a.store.Create(ctx, user)
	if err != nil {
		return model.NoneUser, fmt.Errorf("unable to create: %w", err)
	}

	user.ID = id
	user.Password = ""

	return user, nil
}

// Update User
func (a app) Update(ctx context.Context, user model.User) (model.User, error) {
	if err := a.store.Update(ctx, user); err != nil {
		return user, fmt.Errorf("unable to update: %w", err)
	}

	return user, nil
}

// Delete User
func (a app) Delete(ctx context.Context, user model.User) error {
	if err := a.store.Delete(ctx, user); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return nil
}

func (a app) Check(ctx context.Context, old, new model.User) error {
	output := make([]error, 0)

	user := model.ReadUser(ctx)
	if old != model.NoneUser && user == model.NoneUser {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new == model.NoneUser && !a.auth.IsAuthorized(ctx, user, "admin") {
		output = append(output, errors.New("you must be an admin for deleting"))
	}

	if new == model.NoneUser {
		return httpModel.ConcatError(output)
	}

	if old != model.NoneUser && new != model.NoneUser && !(user.ID == new.ID || a.auth.IsAuthorized(ctx, user, "admin")) {
		output = append(output, errors.New("you're not authorized to interact with other user"))
	}

	if len(strings.TrimSpace(new.Login)) == 0 {
		output = append(output, errors.New("login is required"))
	}

	if old == model.NoneUser && new != model.NoneUser && len(strings.TrimSpace(new.Password)) == 0 {
		output = append(output, errors.New("password is required"))
	}

	return httpModel.ConcatError(output)
}

func (a app) CheckRights(ctx context.Context, id uint64) error {
	user := model.ReadUser(ctx)
	if user == model.NoneUser {
		return httpModel.WrapUnauthorized(errors.New("no user in context"))
	}

	if id != 0 && user.ID == id || a.auth.IsAuthorized(ctx, user, "admin") {
		return nil
	}

	logger.Info("unauthorized access for login=%s", user.Login)

	return httpModel.WrapForbidden(errors.New("unauthorized"))
}
