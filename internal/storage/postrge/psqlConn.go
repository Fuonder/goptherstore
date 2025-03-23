package postrge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"sync"
	"time"
)

var timeouts = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
var maxRetries = len(timeouts)

func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		state := pgErr.SQLState()
		if strings.HasPrefix(state, "08") {
			return true
		}
	}
	return false
}

type Connection struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewConnection(ctx context.Context, settings string) (*Connection, error) {
	var err error
	c := &Connection{}

	logger.Log.Info("Connecting to database")
	c.db, err = sql.Open("pgx", settings)
	if err != nil {
		return &Connection{}, fmt.Errorf("can not connect with database: %v", err)
	}
	logger.Log.Info("Database initial connection successful")
	err = c.ConnectCtx(ctx)
	if err != nil {
		return &Connection{}, fmt.Errorf("access to database: %v", err)
	}

	err = c.MigrateCtx(ctx)
	if err != nil {
		return &Connection{}, fmt.Errorf("%v", err)
	}
	logger.Log.Info("Migration successful")
	return c, nil
}

func (c *Connection) ConnectCtx(ctx context.Context) error {
	var err error
	logger.Log.Info("Checking db accessibility")
	if c.db == nil {
		logger.Log.Warn("no active connection with db")
		return fmt.Errorf("no active connection with db")
	}
	for i := 0; i < maxRetries; i++ {
		err = c.db.PingContext(ctx)
		if err == nil {
			logger.Log.Info("Access - OK")
			return nil
		} else if isConnectionError(err) {
			if i < len(timeouts) {
				logger.Log.Info("can not access database0", zap.Error(err))
				logger.Log.Info("retrying after timeout",
					zap.Duration("timeout", timeouts[i]),
					zap.Int("retry-count", i+1))
				time.Sleep(timeouts[i])
			}
		} else {
			return fmt.Errorf("can not access database1: %v", err)
		}
	}
	return fmt.Errorf("can not access database2: %v", err)
}

func (c *Connection) MigrateCtx(ctx context.Context) error {
	logger.Log.Info("Migrating database")
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, MigrationQuery)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (c *Connection) CheckLoginPresence(ctx context.Context, user storage.MartUser) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var count int
	err := c.db.QueryRowContext(ctx, SearchUserQuery, user.Login).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check login presence: %w", err)
	}
	if count > 0 {
		return storage.ErrUserAlreadyExists
	}
	return nil
}
func (c *Connection) CreateUser(ctx context.Context, newUser storage.MartUser) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// login, pwd, date
	_, err = tx.ExecContext(
		ctx, InsertUserQuery,
		newUser.Login,
		string(hashedPassword),
		newUser.CreatedAt,
	)
	if err != nil {
		return storage.ErrUserCreationFailed
	}
	return tx.Commit()
}

func (c *Connection) ValidateUserCredentials(ctx context.Context, user storage.MartUser) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var hashPassword string
	err := c.db.QueryRowContext(ctx, GetUserPasswordQuery, user.Login).Scan(&hashPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrWrongCredentials
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(user.Password))
	if err != nil {
		return storage.ErrWrongCredentials
	}
	return nil
}

func (c *Connection) WriteNewOrder(ctx context.Context, order storage.MartOrder) error {
	err := c.isOrderExists(ctx, order.OrderID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx, InsertNewOrderQuery,
		order.UserID,
		order.OrderID,
		order.CreatedAt,
		order.Status,
		order.Bonus,
	)
	if err != nil {
		return err
	}
	return tx.Commit()

	// here write to accrual workers jobs channel.

}

func (c *Connection) isOrderExists(ctx context.Context, orderNumber string) error {
	// true - exists (negative case), false - not exists (positive case)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var count int
	err := c.db.QueryRowContext(ctx, SearchOrderByNumberQuery, orderNumber).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check order_number presence: %w", err)
	}
	if count > 0 {
		return storage.ErrOrderAlreadyExists
	}
	return nil
}

func (c *Connection) GetUIDByUsername(ctx context.Context, username string) (int, error) {

	c.mu.RLock()
	defer c.mu.RUnlock()
	var UID int
	err := c.db.QueryRowContext(ctx, GetUIDByUserLoginQuery, username).Scan(&UID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return UID, nil
}

func (c *Connection) Close() error {
	return c.db.Close()
}
