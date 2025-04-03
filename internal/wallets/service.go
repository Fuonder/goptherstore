package wallets

import (
	"context"
	"github.com/Fuonder/goptherstore.git/internal/models"
)

type WalletService interface {
	GetUserBalance(ctx context.Context, UID int) (wallet models.MartUserWallet, err error)
	GetWithdrawals(ctx context.Context, UID int) (withdrawals []models.Withdrawal, err error)
	RegisterWithdraw(ctx context.Context, withdraw models.Withdrawal) error
}
