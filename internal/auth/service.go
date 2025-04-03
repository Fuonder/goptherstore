package auth

import (
	"context"
	"github.com/Fuonder/goptherstore.git/internal/models"
)

type AuthService interface {
	Register(ctx context.Context, newUser models.MartUser) (token string, err error) //+
	Login(ctx context.Context, user models.MartUser) (token string, err error)       //+
	GetJWT(ctx context.Context, login string) (tokenString string, err error)        //+
	ValidateJWT(ctx context.Context, tokenString string) error
	GetUIDFromJWT(ctx context.Context, tokenString string) (int, error)
}
