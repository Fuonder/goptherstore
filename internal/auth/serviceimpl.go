package auth

import (
	"context"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"github.com/Fuonder/goptherstore.git/internal/users"
	"github.com/Fuonder/goptherstore.git/internal/wallets"
	_ "github.com/dgrijalva/jwt-go"
	"time"
)

type AService struct {
	uConn  users.DatabaseUsers
	wConn  wallets.DatabaseWallets
	conn   DatabaseAuth
	secret []byte
}

func NewAService(uConn users.DatabaseUsers, wConn wallets.DatabaseWallets, conn DatabaseAuth, secret []byte) *AService {
	return &AService{
		uConn:  uConn,
		wConn:  wConn,
		conn:   conn,
		secret: secret,
	}
}

func (a *AService) Register(ctx context.Context, newUser models.MartUser) (token string, err error) {
	err = a.uConn.CheckLoginPresence(ctx, newUser)
	if err != nil {
		return "", err
	}

	err = a.uConn.CreateUser(ctx, newUser)
	if err != nil {
		return "", err
	}

	UID, err := a.uConn.GetUIDByUsername(ctx, newUser.Login)
	if err != nil {
		return "", err
	}

	err = a.wConn.CreateUserWallet(ctx, UID)
	if err != nil {
		return "", err
	}

	token, err = a.GetJWT(ctx, newUser.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AService) GetJWT(ctx context.Context, login string) (tokenString string, err error) {
	expirationTime := time.Now().Add(10 * time.Hour)

	claims := &models.Claims{
		Username: login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString(a.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AService) Login(ctx context.Context, user models.MartUser) (token string, err error) {
	err = a.conn.ValidateUserCredentials(ctx, user)
	if err != nil {
		logger.Log.Debug("can not validate user creds")
		return "", err
	}
	token, err = a.GetJWT(ctx, user.Login)
	if err != nil {
		logger.Log.Debug("can not create JWT")
		return "", err
	}
	return token, nil
}

func (a *AService) ValidateJWT(ctx context.Context, tokenString string) error {

	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token uses the correct signing method
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %v", token.Method.Alg())
		}
		return a.secret, nil
	})

	if err != nil || !token.Valid {
		return fmt.Errorf("validation err: %v", err)
	}
	return nil
}

func (a *AService) GetUIDFromJWT(ctx context.Context, tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token uses the correct signing method
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %v", token.Method.Alg())
		}
		return a.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token: %v", err)
	}
	claims, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	UID, err := a.uConn.GetUIDByUsername(ctx, claims.Username)
	if err != nil {
		return 0, fmt.Errorf("error retrieving user ID: %v", err)
	}
	return UID, nil
}
