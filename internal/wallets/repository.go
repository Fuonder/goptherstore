package wallets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	GetWalletByUID = `SELECT balance, total_withdrawn from wallets WHERE user_id = $1;`

	CreateUserWalletQuery = `INSERT INTO wallets (user_id, balance, total_withdrawn, created_at) VALUES ($1, $2, $3, $4);`
	GetBalanceByUID       = `SELECT balance FROM wallets where user_id = $1;`
	InsertWithdraw        = `INSERT INTO withdrawals (user_id, order_number, amount, created_at) VALUES ($1, $2, $3, $4);`
	WithdrawUpdateBalance = `UPDATE wallets SET balance = balance - $1, total_withdrawn = total_withdrawn + $1 WHERE user_id = $2;`
	GetWithdrawalsByUID   = `
						SELECT order_number, amount, created_at 
						FROM withdrawals 
						WHERE user_id = $1 
						ORDER BY created_at DESC;`
	AccrualUpdateBalance = `UPDATE wallets SET balance = balance + $1 WHERE user_id = $2;`
)

type DatabaseWallets interface {
	ProcessWithdraw(ctx context.Context, withdraw models.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, UID int) (withdrawals []models.Withdrawal, err error)
	CreateUserWallet(ctx context.Context, UID int) error
	Accrual(ctx context.Context, value float32, UID int) error
	GetUserWallet(ctx context.Context, UID int) (wallet models.MartUserWallet, err error)
}

type DBWallets struct {
	db *sql.DB
	mu *sync.RWMutex
}

func NewDBWallets(db *sql.DB, mu *sync.RWMutex) (*DBWallets, error) {
	return &DBWallets{db: db, mu: mu}, nil
}

func (w *DBWallets) WriteWithdraw(ctx context.Context, withdraw models.Withdrawal) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx, InsertWithdraw,
		withdraw.UserID,
		withdraw.OrderID,
		withdraw.Amount,
		withdraw.CreatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (w *DBWallets) UpdateWallet(ctx context.Context, value float32, UID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx, WithdrawUpdateBalance,
		value,
		UID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (w *DBWallets) GetUserBalance(ctx context.Context, UID int) (float32, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var balance float32
	err := w.db.QueryRowContext(ctx, GetBalanceByUID, UID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (w *DBWallets) ProcessWithdraw(ctx context.Context, withdraw models.Withdrawal) error {
	//err := w.isUserOrderPresent(ctx, withdraw.OrderID, withdraw.UserID)
	//if err != nil {
	//	return err
	//}
	balance, err := w.GetUserBalance(ctx, withdraw.UserID)
	if err != nil {
		return err
	}
	if balance < withdraw.Amount {
		return models.ErrNotEnoughBonuses
	}
	err = w.WriteWithdraw(ctx, withdraw)
	if err != nil {
		return err
	}
	err = w.UpdateWallet(ctx, withdraw.Amount, withdraw.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (w *DBWallets) GetUserWithdrawals(ctx context.Context, UID int) (withdrawals []models.Withdrawal, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	rows, err := w.db.QueryContext(ctx, GetWithdrawalsByUID, UID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()
	withdrawals = make([]models.Withdrawal, 0)

	for rows.Next() {
		var withdrawal models.Withdrawal
		var amount sql.NullFloat64
		//var createdAt time.Time

		if err := rows.Scan(&withdrawal.OrderID, &amount, &withdrawal.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		//order.CreatedAt = createdAt.Format("2006-01-02T15:04:05-07:00")

		if amount.Valid && amount.Float64 > 0 {
			accrual := float32(amount.Float64)
			withdrawal.Amount = accrual
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}

	logger.Log.Info("RESULT", zap.Any("withdrawals", withdrawals))
	if len(withdrawals) == 0 {
		return nil, models.ErrNoData
	}
	return withdrawals, nil
}

func (w *DBWallets) GetUserWallet(ctx context.Context, UID int) (wallet models.MartUserWallet, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	wallet = models.MartUserWallet{}

	err = w.db.QueryRowContext(ctx, GetWalletByUID, UID).Scan(&wallet.Balance, &wallet.TotalWithdraw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MartUserWallet{}, err
		}
		return models.MartUserWallet{}, fmt.Errorf("failed to get wallet info: %w", err)
	}

	return wallet, nil

}

func (w *DBWallets) CreateUserWallet(ctx context.Context, UID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx, CreateUserWalletQuery,
		UID,
		0,
		0,
		time.Now(),
	)
	if err != nil {
		return err
	}
	return tx.Commit()

}

func (w *DBWallets) Accrual(ctx context.Context, value float32, UID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx, AccrualUpdateBalance,
		value,
		UID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}
