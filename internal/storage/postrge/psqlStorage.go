package postrge

import (
	"context"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type PsqlStorage struct {
	conn   storage.DBConnection
	secret []byte
	//rwMutex sync.RWMutex NOTE: moved to connection
}

func NewPsqlStorage(ctx context.Context, conn storage.DBConnection, key []byte) *PsqlStorage {
	return &PsqlStorage{conn: conn, secret: key}
}

func (p *PsqlStorage) CheckConnection() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = p.conn.ConnectCtx(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *PsqlStorage) Close() error {
	logger.Log.Info("Closing database connection gracefully")
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *PsqlStorage) Register(ctx context.Context, newUser storage.MartUser) (token string, err error) {
	err = p.conn.CheckLoginPresence(ctx, newUser)
	if err != nil {
		return "", err
	}

	err = p.conn.CreateUser(ctx, newUser)
	if err != nil {
		return "", err
	}
	token, err = p.GetJWT(ctx, newUser.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (p *PsqlStorage) GetJWT(ctx context.Context, login string) (tokenString string, err error) {
	expirationTime := time.Now().Add(10 * time.Hour)

	claims := &storage.Claims{
		Username: login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString(p.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (p *PsqlStorage) Login(ctx context.Context, user storage.MartUser) (token string, err error) {
	err = p.conn.ValidateUserCredentials(ctx, user)
	if err != nil {
		logger.Log.Debug("can not validate user creds")
		return "", err
	}
	token, err = p.GetJWT(ctx, user.Login)
	if err != nil {
		logger.Log.Debug("can not create JWT")
		return "", err
	}
	return token, nil
}

func (p *PsqlStorage) RegisterOrder(ctx context.Context, orderNumber string, UID int) error {
	order := storage.MartOrder{
		UserID:    UID,
		OrderID:   orderNumber,
		CreatedAt: time.Now(),
		Status:    "REGISTERED",
		Bonus:     0,
	}
	err := p.conn.WriteNewOrder(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (p *PsqlStorage) GetOrdersByUID(ctx context.Context, UID int) (orders []storage.MartOrder, err error) {
	orders, err = p.conn.GetUserOrders(ctx, UID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (p *PsqlStorage) GetKey() []byte {
	return p.secret
}

func (p *PsqlStorage) GetUID(ctx context.Context, login string) (int, error) {
	UID, err := p.conn.GetUIDByUsername(ctx, login)
	if err != nil {
		return 0, err
	}
	return UID, nil
}

func (p *PsqlStorage) RunWorkers() error {
	return fmt.Errorf("method <RunWorkers>: not implemented")
}
