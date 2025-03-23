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
	"io"
	"net/http"
	"time"
	"unicode"
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
		return
	}

	orderNumberBytes, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		rw.Write([]byte(http.StatusText(http.StatusUnprocessableEntity)))
		return
	}
	logger.Log.Info("GOT ORDER", zap.String("order", string(orderNumberBytes)))

	if ok := isValidLuhn(string(orderNumberBytes)); !ok {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		rw.Write([]byte(http.StatusText(http.StatusUnprocessableEntity)))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	err = h.st.RegisterOrder(ctx, string(orderNumberBytes), UID)
	if err != nil {
		if errors.Is(err, storage.ErrOrderAlreadyExists) {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(http.StatusText(http.StatusOK)))
			return
		} else if errors.Is(err, storage.ErrOrderOfOtherUser) {
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	rw.WriteHeader(http.StatusAccepted)
	rw.Write([]byte(http.StatusText(http.StatusAccepted)))
}

func (h Handlers) GetOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetOrdersHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	orders, err := h.st.GetOrdersByUID(ctx, UID)
	if err != nil {
		if errors.Is(err, storage.ErrNoData) {
			rw.WriteHeader(http.StatusNoContent)
			resp, _ := json.MarshalIndent(http.StatusText(http.StatusNoContent), "", "    ")
			rw.Write(resp)
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	resp, err := json.MarshalIndent(orders, "", "    ")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(resp)
}
func (h Handlers) GetBalanceHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetBalanceHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	wallet, err := h.st.GetUserBalance(ctx, UID)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	resp, err := json.MarshalIndent(wallet, "", "    ")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(resp)
}
func (h Handlers) PostWithdrawHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostWithdrawHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}
	var withdraw storage.Withdrawal

	err := json.NewDecoder(r.Body).Decode(&withdraw)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	withdraw.UserID = UID
	withdraw.CreatedAt = time.Now()

	err = h.st.RegisterWithdraw(ctx, withdraw)
	if err != nil {
		if errors.Is(err, storage.ErrNotEnoughBonuses) {
			rw.WriteHeader(402)
			rw.Write([]byte(err.Error()))
			return
		} else if errors.Is(err, storage.ErrInvalidOrderNumber) {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			rw.Write([]byte(http.StatusText(http.StatusUnprocessableEntity)))
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(http.StatusText(http.StatusOK)))
}
func (h Handlers) GetWithdrawalsHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetWithdrawalsHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	withdrawals, err := h.st.GetWithdrawals(ctx, UID)
	if err != nil {
		if errors.Is(err, storage.ErrNoData) {
			rw.WriteHeader(http.StatusNoContent)
			resp, _ := json.MarshalIndent(http.StatusText(http.StatusNoContent), "", "    ")
			rw.Write(resp)
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.MarshalIndent(http.StatusText(http.StatusInternalServerError), "", "    ")
		rw.Write(resp)
		return
	}

	resp, err := json.MarshalIndent(withdrawals, "", "    ")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	rw.WriteHeader(http.StatusOK)
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

func (h Handlers) getUserID(ctx context.Context, r *http.Request) (int, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not found, return an error
			return 0, fmt.Errorf("cookie not found")
		}
		return 0, fmt.Errorf("error retrieving cookie: %v", err)
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
		return 0, fmt.Errorf("invalid token: %v", err)
	}
	claims, ok := token.Claims.(*storage.Claims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	UID, err := h.st.GetUID(ctx, claims.Username)
	if err != nil {
		return 0, fmt.Errorf("error retrieving user ID: %v", err)
	}
	return UID, nil
}

func isValidLuhn(number string) bool {
	var sum int
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		r := rune(number[i])

		if !unicode.IsDigit(r) {
			return false
		}

		digit := int(r - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
