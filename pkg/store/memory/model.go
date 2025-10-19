package memory

import "github.com/ViBiOh/auth/v3/pkg/model"

type basicUser struct {
	model.User
	password []byte
}
