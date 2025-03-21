package httpserver

import (
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/httpserver/middleware"
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

func (r *RouterObject) GetRouter() (chi.Router, error) {
	if r.chRouter == nil {
		return nil, fmt.Errorf("router not initialized")
	}
	logger.Log.Debug("Configuring Router")
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
	logger.Log.Info("Successfully initialized Router")
	return r.chRouter, nil
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
