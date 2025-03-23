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
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderOfOtherUser   = errors.New("order already registered by other user")
	ErrNoData             = errors.New("no data")
	ErrNotEnoughBonuses   = errors.New("not enough bonuses")
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

var OrderStatus = map[int]string{0: "NEW", 1: "PROCESSING", 2: "INVALID", 3: "PROCESSED"}

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
	ID            int       `json:"-"`
	OwnerID       int       `json:"-"`
	Balance       float32   `json:"current"`
	TotalWithdraw float32   `json:"withdrawn"`
	CreatedAt     time.Time `json:"-"`
}

type MartOrder struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Bonus     float32   `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at,omitempty"`
}

type Withdrawal struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	OrderID   string    `json:"order"`
	Amount    float32   `json:"sum"`
	Status    string    `json:"-"`
	CreatedAt time.Time `json:"processed_at,omitempty"`
}

type DBWriter interface {
	CreateUser(ctx context.Context, newUser MartUser) error
	WriteNewOrder(ctx context.Context, order MartOrder) error
	ProcessWithdraw(ctx context.Context, withdraw Withdrawal) error
	GetUserWithdrawals(ctx context.Context, UID int) (withdrawals []Withdrawal, err error)
	CreateUserWallet(ctx context.Context, login string) error
	Accrual(ctx context.Context, value float32, UID int) error
	UpdateOrder(ctx context.Context, order MartOrder) error
	/*
		1. Записать пользователя
		2. Записать заказ (начисление)
		3. Записать списание
	*/
}

type DBReader interface {
	CheckLoginPresence(ctx context.Context, user MartUser) error
	ValidateUserCredentials(ctx context.Context, user MartUser) error
	GetUIDByUsername(ctx context.Context, username string) (int, error)
	GetUserOrders(ctx context.Context, UID int) ([]MartOrder, error)
	GetUserWallet(ctx context.Context, UID int) (wallet MartUserWallet, err error)
	GetOrderOwner(ctx context.Context, orderNumber string) (UID int, err error)

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

type Storage interface {
	AuthService
	RegisterOrder(ctx context.Context, orderNumber string, UID int) error
	GetUID(ctx context.Context, login string) (int, error)
	GetOrdersByUID(ctx context.Context, UID int) (orders []MartOrder, err error)
	GetUserBalance(ctx context.Context, UID int) (wallet MartUserWallet, err error)
	RegisterWithdraw(ctx context.Context, withdraw Withdrawal) error
	GetWithdrawals(ctx context.Context, UID int) (withdrawals []Withdrawal, err error)
	UpdateOrder(ctx context.Context, order MartOrder) error

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
