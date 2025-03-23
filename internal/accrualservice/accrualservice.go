package accrualservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

type BonusAPIService struct {
	st   storage.Storage
	jobs chan storage.MartOrder
	addr string
}

func NewBonusAPIService(st storage.Storage, jobs chan storage.MartOrder, addr string) *BonusAPIService {
	return &BonusAPIService{st, jobs, addr}
}

func (b *BonusAPIService) Run() error {
	err := b.RunWorkers()
	if err != nil {
		return err
	}
	return nil
}

func (b *BonusAPIService) RunWorkers() error {
	var wg sync.WaitGroup
	g := new(errgroup.Group)

	for i := range 10 {
		wg.Add(1)
		g.Go(func() error {
			err := b.worker(i, b.jobs, &wg)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Debug("workers exited with error", zap.Error(err))
		return fmt.Errorf("method RunWorkers: %v", err)
	}
	return nil
}

func (b *BonusAPIService) worker(idx int, jobs <-chan storage.MartOrder, wg *sync.WaitGroup) error {
	defer wg.Done()
	for job := range jobs {
		logger.Log.Info("processing job", zap.Int("worker", idx))
		logger.Log.Info("JOB", zap.Any("job", job))
		err := b.GetAccrualStatus(job)
		if err != nil {
			logger.Log.Error("error getting accrual status", zap.Error(err))
		}

		//err := middleware.RetryableWorkerHTTPSend(c.Post, "", job, 3)
		//if err != nil {
		//	logger.Log.Debug("sending batch failed", zap.Error(err))
		//	return fmt.Errorf("worker %d: %v", idx, err)
		//}
	}
	return nil
}

func (b *BonusAPIService) GetAccrualStatus(order storage.MartOrder) error {
	retriesCount := 5
	timeouts := make([]time.Duration, retriesCount)
	for i := 0; i < retriesCount; i++ {
		timeouts[i] = time.Duration(2*i+1) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < retriesCount; i++ {
		logger.Log.Info("sending request to accrual")
		responseOrder, err := b.Get(order)
		if err != nil {
			if errors.Is(err, ErrNotRegistered) {
				return nil
			} else if errors.Is(err, ErrToManyRequests) {
				time.Sleep(60 * time.Second)
				i = 0
				continue
			} else if errors.Is(err, ErrInternalServerError) {
				return nil
			} else {
				return err
			}
		}
		if responseOrder.Status == AccrualStatusProcessed || responseOrder.Status == AccrualStatusInvalid {
			err = b.st.UpdateOrder(ctx, responseOrder)
			if err != nil {
				return err
			}
			return nil
		} else {
			i = 0
		}
		if i < len(timeouts) {
			logger.Log.Info("sending failed", zap.Error(err))
			logger.Log.Info("retrying after timeout",
				zap.Duration("timeout", timeouts[i]),
				zap.Int("retry-count", i+1))
			time.Sleep(timeouts[i])
		}
	}
	return ErrCanNotGetAccrualResponse
}

func (b *BonusAPIService) Get(order storage.MartOrder) (storage.MartOrder, error) {
	var err error
	remoteURL := "http://" + b.addr + "/api/orders/" + order.OrderID
	client := resty.New()

	var resp *resty.Response

	body, err := json.Marshal(order)
	if err != nil {
		return storage.MartOrder{}, err
	}

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Get(remoteURL)
	if err != nil {
		return storage.MartOrder{}, err
	}
	if resp.StatusCode() == 200 {
		logger.Log.Info("Accrual response - OK")
		var respBody storage.MartOrder
		err = json.Unmarshal(resp.Body(), &respBody)
		if err != nil {
			return storage.MartOrder{}, err
		}
		return respBody, nil
	} else if resp.StatusCode() == 204 {
		return storage.MartOrder{}, ErrNotRegistered
	} else if resp.StatusCode() == 429 {
		return storage.MartOrder{}, ErrToManyRequests
	} else if resp.StatusCode() == 500 {
		return storage.MartOrder{}, ErrInternalServerError
	} else {
		return storage.MartOrder{}, fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

}
