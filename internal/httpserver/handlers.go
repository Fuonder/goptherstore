package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"net/http"
	"time"
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
	if r.Header.Get("Content-Type") != "application/json" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	var newUser storage.MartUser

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}
	newUser.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := h.st.Register(ctx, newUser)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("User created successfully"))
}
func (h Handlers) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("LoginHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	var user storage.MartUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := h.st.Login(ctx, user)
	if err != nil {
		logger.Log.Debug("error", zap.Error(err))
		if errors.Is(err, storage.ErrWrongCredentials) {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("User login success"))
}
func (h Handlers) PostOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostOrdersHandler called")
	if r.Header.Get("Content-Type") != "text/plain" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) GetOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetOrdersHandler called")
	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(http.StatusNotImplemented)
	resp, _ := json.MarshalIndent(http.StatusText(http.StatusNotImplemented), "", "    ")
	rw.Write(resp)
}
func (h Handlers) GetBalanceHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetBalanceHandler called")
	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(http.StatusNotImplemented)
	resp, _ := json.MarshalIndent(http.StatusText(http.StatusNotImplemented), "", "    ")
	rw.Write(resp)
}
func (h Handlers) PostWithdrawHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostWithdrawHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not Implemented"))
}
func (h Handlers) GetWithdrawalsHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetWithdrawalsHandler called")
	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(http.StatusNotImplemented)
	resp, _ := json.MarshalIndent(http.StatusText(http.StatusNotImplemented), "", "    ")
	rw.Write(resp)
}

func (h Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return logger.HanlderWithLogger(func(rw http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("Auth middleware")

		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie == nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Missing or invalid token"))
			return
		}

		tokenString := cookie.Value
		token, err := jwt.ParseWithClaims(tokenString, &storage.Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token uses the correct signing method
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method %v", token.Method.Alg())
			}
			return h.st.GetKey(), nil
		})

		if err != nil || !token.Valid {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Invalid token"))
			return
		}

		next.ServeHTTP(rw, r)
	})
}
