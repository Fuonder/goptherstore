package httpserver

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

func NewService(APIAddr string, st storage.Storage) (*Service, error) {
	/* TODO:
	   1. init handlers
	   2. create router
	   3. create server
	*/

	h := NewHandlers(st)
	r := NewRouterObject(*h)
	router, err := r.GetRouter()
	if err != nil {
		return nil, err
	}

	service := &Service{
		apiSrv: http.Server{
			Addr:    APIAddr,
			Handler: router,
		},
		st: st,
	}
	return service, nil
}

func (s *Service) Run() error {
	logger.Log.Info("API Listening at",
		zap.String("Addr", s.apiSrv.Addr))
	return s.apiSrv.ListenAndServe()
}
