package storage

import "context"

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
	PingCtx(ctx context.Context) error
}

type AuthService interface {
	Register(ctx context.Context, username string, password string) error
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
