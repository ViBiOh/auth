package servicetest

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/service"
)

var _ service.App = &App{}

// App mocks
type App struct {
	checkErr error

	listUsers []model.User
	listTotal uint
	listErr   error

	getUser model.User
	getErr  error

	createUser model.User
	createErr  error

	updateUser model.User
	updateErr  error

	deleteErr error

	checkRightsErr error
}

// New create new instance
func New() *App {
	return &App{}
}

// SetCheck mocks
func (a *App) SetCheck(err error) *App {
	a.checkErr = err

	return a
}

// SetList mocks
func (a *App) SetList(users []model.User, total uint, err error) *App {
	a.listUsers = users
	a.listTotal = total
	a.listErr = err

	return a
}

// SetGet mocks
func (a *App) SetGet(user model.User, err error) *App {
	a.getUser = user
	a.getErr = err

	return a
}

// SetCreate mocks
func (a *App) SetCreate(user model.User, err error) *App {
	a.createUser = user
	a.createErr = err

	return a
}

// SetUpdate mocks
func (a *App) SetUpdate(user model.User, err error) *App {
	a.updateUser = user
	a.updateErr = err

	return a
}

// SetDelete mocks
func (a *App) SetDelete(err error) *App {
	a.deleteErr = err

	return a
}

// SetCheckRights mocks
func (a *App) SetCheckRights(err error) *App {
	a.checkRightsErr = err

	return a
}

// Check mocks
func (a *App) Check(ctx context.Context, old, new model.User) error {
	return a.checkErr
}

// List mocks
func (a *App) List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]model.User, uint, error) {
	return a.listUsers, a.listTotal, a.listErr
}

// Get mocks
func (a *App) Get(ctx context.Context, ID uint64) (model.User, error) {
	return a.getUser, a.getErr
}

// Create mocks
func (a *App) Create(ctx context.Context, o model.User) (model.User, error) {
	return a.createUser, a.createErr
}

// Update mocks
func (a *App) Update(ctx context.Context, o model.User) (model.User, error) {
	return a.updateUser, a.updateErr
}

// Delete mocks
func (a *App) Delete(ctx context.Context, o model.User) error {
	return a.deleteErr
}

// CheckRights mocks
func (a *App) CheckRights(ctx context.Context, id uint64) error {
	return a.checkRightsErr
}
