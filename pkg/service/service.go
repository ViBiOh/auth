package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var (
	_ crud.Service = &app{}

	// ErrUnknownItemType occurs when item is unknown
	ErrUnknownItemType = errors.New("unknown item type")
)

// App of package
type App interface {
	Unmarsall([]byte) (crud.Item, error)
	Check(crud.Item) []error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]crud.Item, uint, error)
	Get(ctx context.Context, ID uint64) (crud.Item, error)
	Create(ctx context.Context, o crud.Item) (crud.Item, error)
	Update(ctx context.Context, o crud.Item) (crud.Item, error)
	Delete(ctx context.Context, o crud.Item) error
}

type app struct {
	db *sql.DB
}

// New creates new App from Config
func New(db *sql.DB) App {
	return &app{
		db: db,
	}
}

// Unmarsall User
func (a app) Unmarsall(content []byte) (crud.Item, error) {
	var user model.User

	if err := json.Unmarshal(content, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// List Users
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]crud.Item, uint, error) {
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
	item, err := a.get(ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	return &item, nil
}

// Create User
func (a app) Create(ctx context.Context, o crud.Item) (crud.Item, error) {
	user := o.(*model.User)

	id, err := a.create(*user, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create: %w", err)
	}

	o.SetID(id)

	return o, nil
}

// Update User
func (a app) Update(ctx context.Context, o crud.Item) (crud.Item, error) {
	user := o.(*model.User)

	if err := a.update(*user, nil); err != nil {
		return nil, fmt.Errorf("unable to update: %w", err)
	}

	return o, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o crud.Item) (err error) {
	if err := a.delete(*o.(*model.User), nil); err != nil {
		err = fmt.Errorf("unable to delete: %w", err)
	}

	return
}

func (a app) Check(o crud.Item) []error {
	user := o.(*model.User)
	errors := make([]error, 0)

	if strings.TrimSpace(user.Login) == "" {
		errors = append(errors, fmt.Errorf("name is required: %w", crud.ErrInvalid))
	}

	return errors
}
