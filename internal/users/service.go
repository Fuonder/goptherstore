package users

import "context"

type UserService interface {
	GetUID(ctx context.Context, login string) (int, error)
}
