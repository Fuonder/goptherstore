package httpServer

import (
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"net/http"
)

type Handlers struct {
	st storage.Storage
}

func NewHandlers(st storage.Storage) *Handlers {
	return &Handlers{st}
}

func (h Handlers) RootHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("RootHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) RegisterHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("RegisterHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("LoginHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) PostOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostOrdersHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) GetOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetOrdersHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) GetBalanceHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetBalanceHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) PostWithdrawHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostWithdrawHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) GetWithdrawalsHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetWithdrawalsHandler called")
	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
