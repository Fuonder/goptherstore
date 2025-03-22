package storage

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserCreationFailed = errors.New("user creation failed")
	ErrWrongCredentials   = errors.New("wrong credentials")
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type MartUser struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"pwd"`
	CreatedAt time.Time `json:"created_at"`
}

type MartUserWallet struct {
	ID            int       `json:"id"`
	OwnerID       int       `json:"user_id"`
	Balance       int       `json:"balance"`
	TotalWithdraw int       `json:"total_withdraw"`
	CreatedAt     time.Time `json:"created_at"`
}

type MartOrder struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   int       `json:"order_id"`
	Status    string    `json:"status"`
	Bonus     int       `json:"bonus,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Withdrawal struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   int       `json:"order_id"`
	Amount    int       `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type DBWriter interface {
	CreateUser(ctx context.Context, newUser MartUser) error
	/*
		1. Записать пользователя
		2. Записать заказ (начисление)
		3. Записать списание
	*/
}

type DBReader interface {
	CheckLoginPresence(ctx context.Context, user MartUser) error
	ValidateUserCredentials(ctx context.Context, user MartUser) error

	/*
		1. Получить пароль по логину
		2. Получить список всех заказов по пользователю (начислений)
		3. Получить список всех списаний
		4. Получить текущий баланс
		5.
	*/
}

type AuthService interface {
	Register(ctx context.Context, newUser MartUser) (token string, err error)
	Login(ctx context.Context, user MartUser) (token string, err error)
	GetKey() []byte
}

type DBConnection interface {
	DBWriter
	DBReader
	//AuthService
	ConnectCtx(ctx context.Context) error
	MigrateCtx(ctx context.Context) error
	//PingCtx(ctx context.Context) error
	Close() error
}

type AccrualService interface {
	RunWorkers() error
}

type Storage interface {
	AuthService
	AccrualService

	/*
		1. Зарегистрировать пользователя
		2. Аутентифицировать пользователя
		3. Загрузить заказ
		4. Получить заказы
		5. Получить баланс
		6. Загрузить списание
		7. Получить список списаний
	*/
}
