package httpServer

import (
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

type Service struct {
	apiSrv http.Server
	st     storage.Storage
}

func NewService(APIAddr string, st storage.Storage) *Service {
	/* TODO:
	   1. init handlers
	   2. create router
	   3. create server
	*/

	h := NewHandlers(st)
	r := NewRouterObject(*h)

	service := &Service{
		apiSrv: http.Server{
			Addr:    APIAddr,
			Handler: r.GetRouter(),
		},
		st: st,
	}
	return service
}

func (s *Service) Run() error {
	logger.Log.Info("API Listening at",
		zap.String("Addr", s.apiSrv.Addr))
	return s.apiSrv.ListenAndServe()
}
