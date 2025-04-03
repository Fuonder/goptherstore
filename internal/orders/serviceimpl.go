package orders

import (
	"context"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"github.com/Fuonder/goptherstore.git/internal/wallets"
	"go.uber.org/zap"
	"time"
)

type OService struct {
	wConn wallets.DatabaseWallets
	conn  DatabaseOrders
	jobs  chan models.MartOrder
}

func NewOService(conn DatabaseOrders, wConn wallets.DatabaseWallets, jobsCh chan models.MartOrder) *OService {
	return &OService{conn: conn, wConn: wConn, jobs: jobsCh}
}

func (s *OService) RegisterOrder(ctx context.Context, orderNumber string, UID int) error {
	order := models.MartOrder{
		UserID:    UID,
		OrderID:   orderNumber,
		CreatedAt: time.Now(),
		Status:    models.OrderStatusNew,
		Bonus:     0,
	}
	err := s.conn.WriteNewOrder(ctx, order)
	if err != nil {
		return err
	}
	s.jobs <- order
	order.Status = models.OrderStatusProcessing
	err = s.conn.UpdateOrder(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (s *OService) GetOrdersByUID(ctx context.Context, UID int) (orders []models.MartOrder, err error) {
	orders, err = s.conn.GetUserOrders(ctx, UID)
	if err != nil {
		return nil, err
	}
	logger.Log.Info("GOT ORDERS", zap.Any("orders", orders))
	return orders, nil
}

func (s *OService) UpdateOrder(ctx context.Context, order models.MartOrder) error {
	err := s.conn.UpdateOrder(ctx, order)
	if err != nil {
		return err
	}
	//1. update order

	//2. Get user_id from order SearchOrderByNumberQuery
	UID, err := s.conn.GetOrderOwner(ctx, order.OrderID)
	if err != nil {
		return err
	}
	//3. change wallet balance AccrualUpdateBalance
	err = s.wConn.Accrual(ctx, order.Bonus, UID)
	if err != nil {
		return err
	}
	return nil
}
