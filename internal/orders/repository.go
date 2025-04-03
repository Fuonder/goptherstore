package orders

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
	InsertNewOrderQuery = `
							INSERT INTO orders (user_id, order_number, created_at, status, bonus_amount) 
							VALUES ($1, $2, $3, $4, $5);`
	SearchOrderByNumberQuery = `SELECT user_id from orders WHERE order_number = $1;`
	GetOrdersByUID           = `
						SELECT order_number, status, bonus_amount, created_at 
						FROM orders 
						WHERE user_id = $1 
						ORDER BY created_at DESC;`
	UpdateOrder      = `UPDATE orders SET created_at = $1, status = $2 WHERE order_number = $3;`
	UpdateOrderBonus = `UPDATE orders SET bonus_amount = $1 WHERE order_number = $2`
)

type DatabaseOrders interface {
	WriteNewOrder(ctx context.Context, order models.MartOrder) error
	UpdateOrder(ctx context.Context, order models.MartOrder) error
	GetUserOrders(ctx context.Context, UID int) ([]models.MartOrder, error)
	GetOrderOwner(ctx context.Context, orderNumber string) (UID int, err error)
}

type DBOrders struct {
	db *sql.DB
	mu *sync.RWMutex
}

func NewDBOrders(db *sql.DB, mu *sync.RWMutex) (*DBOrders, error) {
	return &DBOrders{db: db, mu: mu}, nil
}

func (o *DBOrders) WriteNewOrder(ctx context.Context, order models.MartOrder) error {
	err := o.isOrderExists(ctx, order.OrderID, order.UserID)
	if err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()

	tx, err := o.db.BeginTx(ctx, nil)
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

func (o *DBOrders) isOrderExists(ctx context.Context, orderNumber string, UID int) error {
	// true - exists (negative case), false - not exists (positive case)
	o.mu.RLock()
	defer o.mu.RUnlock()
	ownerID := 0
	err := o.db.QueryRowContext(ctx, SearchOrderByNumberQuery, orderNumber).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to check order_number presence: %w", err)
	}
	if ownerID != 0 {
		if ownerID == UID {
			return models.ErrOrderAlreadyExists
		}
		return models.ErrOrderOfOtherUser
	}
	return nil
}

func (o *DBOrders) GetUserOrders(ctx context.Context, UID int) ([]models.MartOrder, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	rows, err := o.db.QueryContext(ctx, GetOrdersByUID, UID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()
	orders := make([]models.MartOrder, 0)

	for rows.Next() {
		var order models.MartOrder
		var bonus sql.NullFloat64
		//var createdAt time.Time

		if err := rows.Scan(&order.OrderID, &order.Status, &bonus, &order.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		//order.CreatedAt = createdAt.Format("2006-01-02T15:04:05-07:00")

		if bonus.Valid && bonus.Float64 > 0 {
			accrual := float32(bonus.Float64)
			order.Bonus = accrual
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during roo iteration: %v", err)
	}

	logger.Log.Info("RESULT", zap.Any("orders", orders))
	if len(orders) == 0 {
		return nil, models.ErrNoData
	}
	return orders, nil
}
func (o *DBOrders) GetOrderOwner(ctx context.Context, orderNumber string) (UID int, err error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	ownerID := 0
	err = o.db.QueryRowContext(ctx, SearchOrderByNumberQuery, orderNumber).Scan(&ownerID)
	if err != nil {
		return 0, fmt.Errorf("failed to check order_number presence: %w", err)
	}
	return ownerID, nil
}
func (o *DBOrders) UpdateOrder(ctx context.Context, order models.MartOrder) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(
		ctx,
		UpdateOrder,
		time.Now(),
		order.Status,
		order.OrderID,
	)
	if err != nil {
		return err
	}
	if order.Bonus > 0 {
		_, err = tx.ExecContext(
			ctx,
			UpdateOrderBonus,
			order.Bonus,
			order.OrderID,
		)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
