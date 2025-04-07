package wallets

import (
	"context"
	"github.com/Fuonder/goptherstore.git/internal/models"
)

type WService struct {
	conn DatabaseWallets
}

func NewWService(conn DatabaseWallets) *WService {
	return &WService{conn: conn}
}

func (s *WService) GetUserBalance(ctx context.Context, UID int) (wallet models.MartUserWallet, err error) {
	wallet, err = s.conn.GetUserWallet(ctx, UID)
	if err != nil {
		return models.MartUserWallet{}, err
	}
	return wallet, nil

}

func (s *WService) GetWithdrawals(ctx context.Context, UID int) (withdrawals []models.Withdrawal, err error) {
	withdrawals, err = s.conn.GetUserWithdrawals(ctx, UID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (s *WService) RegisterWithdraw(ctx context.Context, withdraw models.Withdrawal) error {
	return s.conn.ProcessWithdraw(ctx, withdraw)
}
