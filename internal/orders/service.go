package orders

import (
	"context"
	"github.com/Fuonder/goptherstore.git/internal/models"
)

type OrderService interface {
	RegisterOrder(ctx context.Context, orderNumber string, UID int) error
	GetOrdersByUID(ctx context.Context, UID int) (orders []models.MartOrder, err error)
	UpdateOrder(ctx context.Context, order models.MartOrder) error
}
