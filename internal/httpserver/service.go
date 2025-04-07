package httpserver

import (
	"github.com/Fuonder/goptherstore.git/internal/dbservices"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

type Service struct {
	apiSrv     http.Server
	DBServices *dbservices.DatabaseServices
}

func NewService(APIAddr string, DBServices *dbservices.DatabaseServices) (*Service, error) {

	h := NewHandlers(DBServices)
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
		DBServices: DBServices,
	}
	return service, nil
}

func (s *Service) Run() error {
	logger.Log.Info("API Listening at",
		zap.String("Addr", s.apiSrv.Addr))
	return s.apiSrv.ListenAndServe()
}
