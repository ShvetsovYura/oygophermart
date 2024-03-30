package accrualagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/ShvetsovYura/oygophermart/internal/utils"
	"github.com/dghubble/sling"
	"golang.org/x/sync/errgroup"
)

type Saver interface {
	GetOrdersToAccrualProcess(ctx context.Context) ([]models.OrderModel, error)
	UpdateOrdersStatus(context.Context, ...models.AccrualResult) error
}

type TaskResult struct {
	data *models.AccrualResult
	err  error
}

type AccrualAgent struct {
	accrualLink       string
	saveService       Saver
	workers           int
	taskResultCh      chan TaskResult
	accrualTicker     *time.Ticker
	defaultRetryAfter int
}

func NewAccrualAgent(accrualAddr string, service Saver, interval uint) *AccrualAgent {
	instance := &AccrualAgent{
		accrualLink:       accrualAddr,
		saveService:       service,
		workers:           3,
		taskResultCh:      make(chan TaskResult, 1000),
		accrualTicker:     time.NewTicker(time.Duration(interval) * time.Second),
		defaultRetryAfter: 5,
	}

	return instance
}

var ErrOrderNotRegistered = errors.New("order not registered in accrual system")
var ErrTooManyIntegrationRequests = errors.New("too many requests to accrual service")

type TooManyRequestsError struct {
	RetryAfter int
	Err        error
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("%v retry after: %d", e.Err, e.RetryAfter)
}
func (e *TooManyRequestsError) Unwrap() error {
	return e.Err
}
func NewTooManyRequestsError(err error, retryAfter int) *TooManyRequestsError {
	return &TooManyRequestsError{
		RetryAfter: retryAfter,
		Err:        err,
	}
}

func (a *AccrualAgent) Start(ctx context.Context) {
	go a.startProccessFlushAccrual(ctx)
	go a.startAccrualProcess(ctx)
	<-ctx.Done()
}

func (a *AccrualAgent) getAccrualStatusRequest(orderID string) (*models.AccrualResult, error) {
	s := sling.New().Base(a.accrualLink).Set("User-Agent", "OyGopherMart client")
	r, err := s.New().Get("/api/orders/" + orderID).Request()

	if err != nil {
		logger.Log.Debug("error on create request")
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Log.Debug("error on Do request")
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		logger.Log.Debug("Succes get status")
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result models.AccrualResult
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}

		return &result, nil
	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		var retryafter int
		retryHeder := resp.Header.Get("Retry-After")
		retryafter, err := strconv.Atoi(retryHeder)
		if err != nil {
			retryafter = a.defaultRetryAfter
		}
		logger.Log.Debugf("Too many requests, retryafter: %d", retryafter)
		return nil, NewTooManyRequestsError(ErrTooManyIntegrationRequests, retryafter)
	case http.StatusInternalServerError:
		return nil, errors.New("server error on make request")
	default:
		return nil, errors.New("not success")
	}

}

func (a *AccrualAgent) accrualWorker(orders <-chan string) error {
	for oid := range orders {
		resp, err := a.getAccrualStatusRequest(oid)
		if err != nil {
			if errors.Is(err, ErrTooManyIntegrationRequests) {
				return err
			}
		}
		a.taskResultCh <- TaskResult{
			data: resp,
			err:  err,
		}
	}
	return nil
}

func (a *AccrualAgent) getAccrualInfoTick() {
	eg := &errgroup.Group{}
	logger.Log.Debug("start tick func")
	// взять заказы из БД и отправить номера в WorkerPool
	dbOrders, err := a.saveService.GetOrdersToAccrualProcess(context.Background())
	if err != nil {
		logger.Log.Fatal(err)
	}
	// номера заказов в канал
	ordersToCheckCh := make(chan string, len(dbOrders))

	for w := 0; w < a.workers; w++ {
		eg.Go(func() error {
			err := a.accrualWorker(ordersToCheckCh)
			if err != nil {
				return err
			}
			return nil
		})
	}

	// отправка заказов в канал -
	for _, order := range dbOrders {
		ordersToCheckCh <- order.ID
	}
	close(ordersToCheckCh)

	if err = eg.Wait(); err != nil {
		var target *TooManyRequestsError
		if errors.As(err, &target) {

			logger.Log.Debug("Too many requests")
			a.accrualTicker.Reset(time.Duration(target.RetryAfter * int(time.Second)))
		}
	}
}

func (a *AccrualAgent) startProccessFlushAccrual(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	var records []models.AccrualResult
	for {
		select {
		case rec := <-a.taskResultCh:
			if rec.data != nil {
				if !utils.Contains(*rec.data, records) {
					records = append(records, *rec.data)
				}
			}
		case <-ticker.C:
			if len(records) == 0 {
				logger.Log.Debug("No record to write")
				continue
			}
			err := a.saveService.UpdateOrdersStatus(ctx, records...)
			if err != nil {
				logger.Log.Debug(err)
				continue
			}
			records = nil
		case <-ctx.Done():
			logger.Log.Debug("flush accrual ended")
			return
		}
	}
}

func (a *AccrualAgent) startAccrualProcess(ctx context.Context) {
	for {
		select {
		case <-a.accrualTicker.C:
			a.getAccrualInfoTick()
		case <-ctx.Done():
			logger.Log.Debug("accrual process ended")
			return
		}
	}
}
