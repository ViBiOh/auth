package storetest

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
)

var _ store.UserStorage = &App{}

// App mock app
type App struct {
	doAtomicErr error

	getUser model.User
	getErr  error

	createID  uint64
	createErr error

	updateErr error

	deleteErr error
}

// New creates raw mock
func New() *App {
	return &App{}
}

// SetDoAtomic mocks
func (a *App) SetDoAtomic(err error) *App {
	a.doAtomicErr = err
	return a
}

// SetGet mocks
func (a *App) SetGet(user model.User, err error) *App {
	a.getUser = user
	a.getErr = err

	return a
}

// SetCreate mocks
func (a *App) SetCreate(id uint64, err error) *App {
	a.createID = id
	a.createErr = err

	return a
}

// SetUpdate mocks
func (a *App) SetUpdate(err error) *App {
	a.updateErr = err
	return a
}

// SetDelete mocks
func (a *App) SetDelete(err error) *App {
	a.deleteErr = err
	return a
}

// DoAtomic mocks
func (a *App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return a.doAtomicErr
	}

	return action(ctx)
}

// Get mocks
func (a *App) Get(_ context.Context, _ uint64) (model.User, error) {
	return a.getUser, a.getErr
}

// Create mocks
func (a *App) Create(_ context.Context, _ model.User) (uint64, error) {
	return a.createID, a.createErr
}

// Update mocks
func (a *App) Update(_ context.Context, _ model.User) error {
	return a.updateErr
}

// Delete mocks
func (a *App) Delete(_ context.Context, _ model.User) error {
	return a.deleteErr
}
