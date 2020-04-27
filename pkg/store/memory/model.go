package memory

import "github.com/ViBiOh/auth/v2/pkg/model"

type basicUser struct {
	model.User
	password []byte
}
