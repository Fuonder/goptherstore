package httpServer

import (
	"github.com/Fuonder/goptherstore.git/internal/httpServer/middleware"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/go-chi/chi/v5"
)

type RouterObject struct {
	h        Handlers
	chRouter chi.Router
}

func NewRouterObject(h Handlers) *RouterObject {
	return &RouterObject{h: h, chRouter: chi.NewRouter()}
}

func (r *RouterObject) GetRouter() chi.Router {
	logger.Log.Debug("Creating Router")
	r.chRouter.Route("/api/user", func(router chi.Router) {
		router.Route("/register", func(router chi.Router) {
			router.Post("/", logger.HanlderWithLogger(r.h.RegisterHandler))
		})
		router.Route("/login", func(router chi.Router) {
			router.Post("/", logger.HanlderWithLogger(r.h.LoginHandler))
		})

		router.Route("/orders", func(router chi.Router) {
			router.Use(middleware.Auth)
			router.Post("/", logger.HanlderWithLogger(r.h.PostOrdersHandler))
			router.Get("/", logger.HanlderWithLogger(r.h.GetOrdersHandler))
		})
		router.Route("/balance", func(router chi.Router) {
			router.Use(middleware.Auth)
			router.Get("/", logger.HanlderWithLogger(r.h.GetBalanceHandler))
			router.Post("/withdraw", logger.HanlderWithLogger(r.h.PostWithdrawHandler))
		})
		router.Route("/withdrawals", func(router chi.Router) {
			router.Use(middleware.Auth)
			router.Post("/", logger.HanlderWithLogger(r.h.GetWithdrawalsHandler))
		})
	})
	return r.chRouter
}

/*
POST /api/user/register
POST /api/user/login
POST /api/user/orders
GET /api/user/orders
GET /api/user/balance
POST /api/user/balance/withdraw
GET /api/user/withdrawals
*/
