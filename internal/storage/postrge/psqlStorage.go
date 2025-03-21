package postrge

import (
	"context"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"sync"
	"time"
)

type PsqlStorage struct {
	conn    storage.DBConnection
	rwMutex sync.RWMutex
}

func NewPsqlStorage(ctx context.Context, conn storage.DBConnection) *PsqlStorage {
	return &PsqlStorage{conn: conn}
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

func (p *PsqlStorage) Register(ctx context.Context, username string, password string) (token string, err error) {
	return "", fmt.Errorf("method <Register>: not implemented")
}
func (p *PsqlStorage) Login(ctx context.Context, username string, password string) (token string, err error) {
	return "", fmt.Errorf("method <Register>: not implemented")
}
func (p *PsqlStorage) RunWorkers() error {
	return fmt.Errorf("method <RunWorkers>: not implemented")
}
