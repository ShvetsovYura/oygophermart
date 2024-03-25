package accrualagent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/ShvetsovYura/oygophermart/internal/utils"
	"github.com/dghubble/sling"
)

type Saver interface {
	GetOrdersToAccrualProcess(ctx context.Context) ([]models.OrderModel, error)
	UpdateOrdersStatus(context.Context, ...models.AccrualResult) error
}

type AccrualAgent struct {
	accrualLink string
	saveService Saver
	njobs       int
	workers     int
	interval    time.Duration
	recordsCh   chan models.AccrualResult
}

func NewAccrualAgent(accrualAddr string, service Saver, njobs int, interval time.Duration) *AccrualAgent {
	instance := &AccrualAgent{
		accrualLink: accrualAddr,
		saveService: service,
		njobs:       njobs,
		interval:    interval,
		workers:     3,
		recordsCh:   make(chan models.AccrualResult, 100),
	}

	return instance
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

	if resp.StatusCode != http.StatusOK {
		// logger.Log.Debugf("not success status: %v", resp.StatusCode)
		return nil, errors.New("not success")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Debug("error on read response body")
		return nil, err
	}
	defer resp.Body.Close()

	var result models.AccrualResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.Log.Debugf("error on unmarshall agent: %e", err)
		return nil, err
	}
	logger.Log.Debug(resp.StatusCode)
	logger.Log.Debugf("result: %v", result)
	return &result, nil
}

func (a *AccrualAgent) accrualWorker(orders <-chan string) {
	for oid := range orders {
		resp, _ := a.getAccrualStatusRequest(oid)
		if resp != nil {
			a.recordsCh <- *resp
		}
	}
}

func (a *AccrualAgent) getAccrualInfoTick() {
	logger.Log.Debug("start tick func")
	// взять заказы из БД и отправить номера в WorkerPool
	dbOrders, err := a.saveService.GetOrdersToAccrualProcess(context.Background())
	if err != nil {
		logger.Log.Fatal(err)
	}
	// номера заказов в канал
	ordersToCheckCh := make(chan string, a.njobs)
	// logger.Log.Debug(dbOrders)
	for w := 1; w <= a.workers; w++ {
		go a.accrualWorker(ordersToCheckCh)
	}

	// отправка заказов в канал -
	for _, order := range dbOrders {
		ordersToCheckCh <- order.ID
	}
	close(ordersToCheckCh)
}

func (a *AccrualAgent) startProccessFlushAccrual(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	var records []models.AccrualResult
	for {
		select {
		case rec := <-a.recordsCh:
			if !utils.Contains(rec, records) {
				records = append(records, rec)
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
	t := time.NewTicker(a.interval)
	for {
		select {
		case <-t.C:
			logger.Log.Debug("tick info from Accrual service")
			a.getAccrualInfoTick()
		case <-ctx.Done():
			logger.Log.Debug("accrual process ended")
			return
		}
	}
}
