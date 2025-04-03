package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/auth"
	"github.com/Fuonder/goptherstore.git/internal/dbservices"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/models"
	"github.com/Fuonder/goptherstore.git/internal/orders"
	"github.com/Fuonder/goptherstore.git/internal/users"
	"github.com/Fuonder/goptherstore.git/internal/wallets"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
	"unicode"
)

type Handlers struct {
	userSrv   users.UserService
	walletSrv wallets.WalletService
	orderSrv  orders.OrderService
	authSrv   auth.AuthService
}

func NewHandlers(DBServices *dbservices.DatabaseServices) *Handlers {
	return &Handlers{userSrv: DBServices.UserSrv,
		walletSrv: DBServices.WalletSrv,
		orderSrv:  DBServices.OrderSrv,
		authSrv:   DBServices.AuthSrv}
}

func (h Handlers) RootHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("RootHandler called")
	SendResponse(rw, http.StatusNotImplemented, []byte("Not implemented"))
}

func (h Handlers) RegisterHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("RegisterHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}

	var newUser models.MartUser

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}
	newUser.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := h.authSrv.Register(ctx, newUser)
	if err != nil {
		if errors.Is(err, models.ErrUserAlreadyExists) {
			SendResponse(rw, http.StatusConflict, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	SendResponse(rw, http.StatusOK, []byte("User created successfully"))
}
func (h Handlers) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("LoginHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}

	var user models.MartUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := h.authSrv.Login(ctx, user)
	if err != nil {
		logger.Log.Debug("error", zap.Error(err))
		if errors.Is(err, models.ErrWrongCredentials) {
			SendResponse(rw, http.StatusUnauthorized, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	SendResponse(rw, http.StatusOK, []byte("User login success"))
}
func (h Handlers) PostOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostOrdersHandler called")
	if r.Header.Get("Content-Type") != "text/plain" {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}

	orderNumberBytes, err := io.ReadAll(r.Body)
	if err != nil {
		SendResponse(rw, http.StatusUnprocessableEntity, []byte{})
		return
	}
	logger.Log.Info("GOT ORDER", zap.String("order", string(orderNumberBytes)))

	if ok := isValidLuhn(string(orderNumberBytes)); !ok {
		SendResponse(rw, http.StatusUnprocessableEntity, []byte{})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	err = h.orderSrv.RegisterOrder(ctx, string(orderNumberBytes), UID)
	if err != nil {
		if errors.Is(err, models.ErrOrderAlreadyExists) {
			SendResponse(rw, http.StatusOK, []byte{})
			return
		} else if errors.Is(err, models.ErrOrderOfOtherUser) {
			SendResponse(rw, http.StatusConflict, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	SendResponse(rw, http.StatusAccepted, []byte{})
}

func (h Handlers) GetOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetOrdersHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	ord, err := h.orderSrv.GetOrdersByUID(ctx, UID)
	if err != nil {
		if errors.Is(err, models.ErrNoData) {
			SendResponse(rw, http.StatusNoContent, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	resp, err := json.MarshalIndent(ord, "", "    ")
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	SendResponse(rw, http.StatusOK, resp)
}
func (h Handlers) GetBalanceHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetBalanceHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	wallet, err := h.walletSrv.GetUserBalance(ctx, UID)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	resp, err := json.MarshalIndent(wallet, "", "    ")
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	SendResponse(rw, http.StatusOK, resp)
}
func (h Handlers) PostWithdrawHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("PostWithdrawHandler called")
	if r.Header.Get("Content-Type") != "application/json" {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}
	var withdraw models.Withdrawal

	err := json.NewDecoder(r.Body).Decode(&withdraw)
	if err != nil {
		SendResponse(rw, http.StatusBadRequest, []byte{})
		return
	}
	if ok := isValidLuhn(withdraw.OrderID); !ok {
		SendResponse(rw, http.StatusUnprocessableEntity, []byte{})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	withdraw.UserID = UID
	withdraw.CreatedAt = time.Now()

	err = h.walletSrv.RegisterWithdraw(ctx, withdraw)
	if err != nil {
		if errors.Is(err, models.ErrNotEnoughBonuses) {
			SendResponse(rw, 402, []byte(err.Error()))
			return
		} else if errors.Is(err, models.ErrInvalidOrderNumber) {
			SendResponse(rw, http.StatusUnprocessableEntity, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	SendResponse(rw, http.StatusOK, []byte{})
}
func (h Handlers) GetWithdrawalsHandler(rw http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("GetWithdrawalsHandler called")
	rw.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UID, err := h.getUserID(ctx, r)
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	withdrawals, err := h.walletSrv.GetWithdrawals(ctx, UID)
	if err != nil {
		if errors.Is(err, models.ErrNoData) {
			SendResponse(rw, http.StatusNoContent, []byte{})
			return
		}
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}

	resp, err := json.MarshalIndent(withdrawals, "", "    ")
	if err != nil {
		SendResponse(rw, http.StatusInternalServerError, []byte{})
		return
	}
	SendResponse(rw, http.StatusOK, resp)
}

func (h Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return logger.HanlderWithLogger(func(rw http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("Auth middleware")

		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie == nil {
			SendResponse(rw, http.StatusUnauthorized, []byte("Missing or invalid token"))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tokenString := cookie.Value
		err = h.authSrv.ValidateJWT(ctx, tokenString)
		if err != nil {
			SendResponse(rw, http.StatusUnauthorized, []byte("Invalid token"))
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func (h Handlers) getUserID(ctx context.Context, r *http.Request) (int, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// If the cookie is not found, return an error
			return 0, fmt.Errorf("cookie not found")
		}
		return 0, fmt.Errorf("error retrieving cookie: %v", err)
	}
	tokenString := cookie.Value
	UID, err := h.authSrv.GetUIDFromJWT(ctx, tokenString)
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

func SendResponse(rw http.ResponseWriter, status int, message []byte) {
	rw.WriteHeader(status)
	if len(message) == 0 {
		_, _ = rw.Write([]byte(http.StatusText(status)))
		return
	}
	_, _ = rw.Write(message)
}
