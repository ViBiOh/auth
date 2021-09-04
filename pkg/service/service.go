package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
)

// App of package
type App struct {
	storeApp auth.Storage
	authApp  auth.Provider
}

// New creates new App from Config
func New(storeApp auth.Storage, authApp auth.Provider) App {
	return App{
		storeApp: storeApp,
		authApp:  authApp,
	}
}

// Get User
func (a App) Get(ctx context.Context, ID uint64) (model.User, error) {
	if err := a.CheckRights(ctx, ID); err != nil {
		return model.User{}, err
	}

	item, err := a.storeApp.Get(ctx, ID)
	if err != nil {
		return model.User{}, fmt.Errorf("unable to get: %w", err)
	}

	if item.IsZero() {
		return model.User{}, httpModel.WrapNotFound(errors.New("user not found"))
	}

	return item, nil
}

// Create User
func (a App) Create(ctx context.Context, user model.User) (model.User, error) {
	id, err := a.storeApp.Create(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("unable to create: %w", err)
	}

	user.ID = id
	user.Password = ""

	return user, nil
}

// Update User
func (a App) Update(ctx context.Context, user model.User) (model.User, error) {
	if err := a.storeApp.Update(ctx, user); err != nil {
		return user, fmt.Errorf("unable to update: %w", err)
	}

	return user, nil
}

// Delete User
func (a App) Delete(ctx context.Context, user model.User) error {
	if err := a.storeApp.Delete(ctx, user); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return nil
}

// Check user values
func (a App) Check(ctx context.Context, old, new model.User) error {
	output := make([]error, 0)

	user := model.ReadUser(ctx)
	if !old.IsZero() && user.IsZero() {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new.IsZero() && !a.authApp.IsAuthorized(ctx, user, "admin") {
		output = append(output, errors.New("you must be an admin for deleting"))
	}

	if new.IsZero() {
		return httpModel.ConcatError(output)
	}

	if !old.IsZero() && !new.IsZero() && !(user.ID == new.ID || a.authApp.IsAuthorized(ctx, user, "admin")) {
		output = append(output, errors.New("you're not authorized to interact with other user"))
	}

	if len(new.Login) == 0 {
		output = append(output, errors.New("login is required"))
	}

	if old.IsZero() && !new.IsZero() && len(new.Password) == 0 {
		output = append(output, errors.New("password is required"))
	}

	return httpModel.ConcatError(output)
}

// CheckRights of user ID
func (a App) CheckRights(ctx context.Context, id uint64) error {
	user := model.ReadUser(ctx)
	if user.IsZero() {
		return httpModel.WrapUnauthorized(errors.New("no user in context"))
	}

	if id != 0 && user.ID == id || a.authApp.IsAuthorized(ctx, user, "admin") {
		return nil
	}

	logger.WithField("login", user.Login).Info("unauthorized access")

	return httpModel.WrapForbidden(errors.New("unauthorized"))
}
