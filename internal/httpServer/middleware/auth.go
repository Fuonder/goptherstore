package middleware

import (
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"net/http"
)

func Auth(next http.Handler) http.Handler {
	return logger.HanlderWithLogger(func(rw http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("Auth middleware")

		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		
		//// TODO: JWT Validation
		//next.ServeHTTP(rw, r)
	})
}
