package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
)

type Service struct {
	storeService auth.Storage
	authService  auth.Provider
}

func New(storeService auth.Storage, authService auth.Provider) Service {
	return Service{
		storeService: storeService,
		authService:  authService,
	}
}

func (s Service) Get(ctx context.Context, ID uint64) (model.User, error) {
	if err := s.CheckRights(ctx, ID); err != nil {
		return model.User{}, err
	}

	item, err := s.storeService.Get(ctx, ID)
	if err != nil {
		return model.User{}, fmt.Errorf("get: %w", err)
	}

	if item.IsZero() {
		return model.User{}, httpModel.WrapNotFound(errors.New("user not found"))
	}

	return item, nil
}

func (s Service) Create(ctx context.Context, user model.User) (model.User, error) {
	id, err := s.storeService.Create(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("create: %w", err)
	}

	user.ID = id
	user.Password = ""

	return user, nil
}

func (s Service) Update(ctx context.Context, user model.User) (model.User, error) {
	if err := s.storeService.Update(ctx, user); err != nil {
		return user, fmt.Errorf("update: %w", err)
	}

	return user, nil
}

func (s Service) Delete(ctx context.Context, user model.User) error {
	if err := s.storeService.Delete(ctx, user); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (s Service) Check(ctx context.Context, old, new model.User) error {
	var output []error

	user := model.ReadUser(ctx)
	if !old.IsZero() && user.IsZero() {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new.IsZero() && !s.authService.IsAuthorized(ctx, user, "admin") {
		output = append(output, errors.New("you must be an admin for deleting"))
	}

	if new.IsZero() {
		return httpModel.ConcatError(output)
	}

	if !old.IsZero() && !new.IsZero() && !(user.ID == new.ID || s.authService.IsAuthorized(ctx, user, "admin")) {
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

func (s Service) CheckRights(ctx context.Context, id uint64) error {
	user := model.ReadUser(ctx)
	if user.IsZero() {
		return httpModel.WrapUnauthorized(errors.New("no user in context"))
	}

	if id != 0 && user.ID == id || s.authService.IsAuthorized(ctx, user, "admin") {
		return nil
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "unauthorized access", slog.String("login", user.Login))

	return httpModel.WrapForbidden(errors.New("unauthorized"))
}
