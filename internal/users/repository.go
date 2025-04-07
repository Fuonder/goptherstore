package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

const (
	InsertUserQuery = `
						INSERT INTO users (login, password_hash, created_at) 
						VALUES ($1, $2, $3);
						`
	SearchUserQuery        = `SELECT COUNT(*) FROM users WHERE login = $1;`
	GetUIDByUserLoginQuery = `SELECT id FROM users WHERE login = $1;`
)

type DatabaseUsers interface {
	CreateUser(ctx context.Context, newUser models.MartUser) error
	CheckLoginPresence(ctx context.Context, user models.MartUser) error
	GetUIDByUsername(ctx context.Context, username string) (int, error)
}

type DBUsers struct {
	db *sql.DB
	mu *sync.RWMutex
}

func NewDBUsers(db *sql.DB, mu *sync.RWMutex) (*DBUsers, error) {
	return &DBUsers{db: db, mu: mu}, nil
}

func (u *DBUsers) CreateUser(ctx context.Context, newUser models.MartUser) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	tx, err := u.db.BeginTx(ctx, nil)
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
		return models.ErrUserCreationFailed
	}

	return tx.Commit()
}

func (u *DBUsers) CheckLoginPresence(ctx context.Context, user models.MartUser) error {
	u.mu.RLock()
	defer u.mu.RUnlock()
	var count int
	err := u.db.QueryRowContext(ctx, SearchUserQuery, user.Login).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check login presence: %w", err)
	}
	if count > 0 {
		return models.ErrUserAlreadyExists
	}
	return nil
}

func (u *DBUsers) GetUIDByUsername(ctx context.Context, username string) (int, error) {

	u.mu.RLock()
	defer u.mu.RUnlock()
	var UID int
	err := u.db.QueryRowContext(ctx, GetUIDByUserLoginQuery, username).Scan(&UID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return UID, nil
}
