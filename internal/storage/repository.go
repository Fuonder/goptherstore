package storage

import (
	"context"
	"time"
)

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
	/*
		1. Записать пользователя
		2. Записать заказ (начисление)
		3. Записать списание
	*/
}

type DBReader interface {
	/*
		1. Получить пароль по логину
		2. Получить список всех заказов по пользователю (начислений)
		3. Получить список всех списаний
		4. Получить текущий баланс
		5.
	*/
}

type DBConnection interface {
	DBWriter
	DBReader
	ConnectCtx(ctx context.Context) error
	MigrateCtx(ctx context.Context) error
	//PingCtx(ctx context.Context) error
	Close() error
}

type AuthService interface {
	Register(ctx context.Context, username string, password string) (token string, err error)
	Login(ctx context.Context, username string, password string) (token string, err error)
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
