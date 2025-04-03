package postrge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

func Timeouts() [3]time.Duration {
	return [3]time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
}

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

	timeouts := Timeouts()
	maxRetries := len(timeouts)

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

func (c *Connection) GetDBInstance(ctx context.Context) (*sql.DB, *sync.RWMutex, error) {
	if c.db == nil {
		return nil, nil, fmt.Errorf("no connection present")
	}
	err := c.ConnectCtx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("access to database: %v", err)
	}
	return c.db, &c.mu, nil
}

func (c *Connection) Close() error {
	return c.db.Close()
}
