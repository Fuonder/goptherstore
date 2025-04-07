package dbservices

import (
	"database/sql"
	"github.com/Fuonder/goptherstore.git/internal/auth"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"github.com/Fuonder/goptherstore.git/internal/orders"
	"github.com/Fuonder/goptherstore.git/internal/users"
	"github.com/Fuonder/goptherstore.git/internal/wallets"
	"sync"
)

type DatabaseServices struct {
	UserSrv   users.UserService
	WalletSrv wallets.WalletService
	OrderSrv  orders.OrderService
	AuthSrv   auth.AuthService
}

func NewDatabaseServices(jobsCh chan models.MartOrder, secret []byte, db *sql.DB, mu *sync.RWMutex) (*DatabaseServices, error) {
	s := &DatabaseServices{}

	// user -> wallet -> order -> auth

	DBUsers, err := users.NewDBUsers(db, mu)
	if err != nil {
		return s, err
	}

	s.UserSrv = users.NewUService(DBUsers)

	DBWallets, err := wallets.NewDBWallets(db, mu)
	if err != nil {
		return s, err
	}

	s.WalletSrv = wallets.NewWService(DBWallets)

	DBOrders, err := orders.NewDBOrders(db, mu)
	if err != nil {
		return s, err
	}

	s.OrderSrv = orders.NewOService(DBOrders, DBWallets, jobsCh)

	DBAuth, err := auth.NewDBAuth(db, mu)
	if err != nil {
		return s, err
	}

	s.AuthSrv = auth.NewAService(DBUsers, DBWallets, DBAuth, secret)

	return s, nil
}
