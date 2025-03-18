package http

import (
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"net/http"
)

type Service struct {
	srv http.Server
	st  storage.Storage
}
