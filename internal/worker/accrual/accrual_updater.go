package accrual

import (
	"context"
	"sync"
	"time"

	"errors"

	"github.com/alxaxenov/ya-gophermart/internal/client/accrual"
	d_o "github.com/alxaxenov/ya-gophermart/internal/domain/order"
	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
)

type Updater struct {
	repo          gophermartRepo
	accrualClient accrualClient
	mu            sync.RWMutex
	wg            sync.WaitGroup
	waitRequest   time.Time
	ch            chan model.OrderToProcess
	workersNum    int
}

type gophermartRepo interface {
	GetOrdersToProcess(context.Context, int) (*[]model.OrderToProcess, error)
	UpdateOrderStatus(context.Context, string, d_o.Status) error
	UpdateOrderAccrual(context.Context, string, int, d_o.Status, float64) error
}

type accrualClient interface {
	GetOrderInfo(string) (*accrual.OrderInfo, error)
}

func NewUpdater(ctx context.Context, repo gophermartRepo, accrualClient accrualClient, workersNum int) *Updater {
	ch := make(chan model.OrderToProcess, workersNum*2)
	wg := sync.WaitGroup{}
	wait := time.Time{}
	mu := sync.RWMutex{}
	u := &Updater{repo, accrualClient, mu, wg, wait, ch, workersNum}
	for i := 0; i < workersNum; i++ {
		go u.worker(ctx)
	}
	return u
}

func (u *Updater) Run(ctx context.Context) {
	defer close(u.ch)
	wait := time.Now().Add(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(wait)):
			wait = time.Now().Add(10 * time.Second)
			orders, err := u.repo.GetOrdersToProcess(ctx, u.workersNum*2)
			if err != nil {
				logger.Logger.Errorf("Updater failed to get orders from db: %v", err)
			}
			if len(*orders) == 0 {
				continue
			}
			for _, order := range *orders {
				u.ch <- order
			}
			u.wg.Wait()
		}
	}
}

func (u *Updater) worker(ctx context.Context) {
	for orderNumber := range u.ch {
		u.innerWorker(ctx, &orderNumber)
	}
}

func (u *Updater) innerWorker(ctx context.Context, order *model.OrderToProcess) {
	u.wg.Add(1)
	defer u.wg.Done()

	if order.Status == d_o.NEW {
		err := u.repo.UpdateOrderStatus(ctx, order.Number, d_o.PROCESSING)
		if err != nil {
			logger.Logger.Errorf("Updater failed to update order initial status: %v", err)
			return
		}
	}

	ctx_done := u.waitUntilNextRequestAllowed(ctx)
	if ctx_done {
		return
	}

	info, err := u.accrualClient.GetOrderInfo(order.Number)
	if err != nil {
		var rateLimit *accrual.RateLimitError
		switch {
		case errors.Is(err, accrual.OrderNotRegisteredError):
			return
		case errors.As(err, &rateLimit):
			logger.Logger.Infof("Updater accrual timeout responce, wait for: %d", rateLimit.Timeout)
			u.updateNextRequestAllowed(rateLimit.Timeout)
			return
		default:
			logger.Logger.Errorf("Updater failed to get accrual order info: %v", err)
			return
		}
	}

	if !info.Status.IsFinal() {
		return
	}
	if info.Accrual == 0 {
		err = u.repo.UpdateOrderStatus(ctx, order.Number, info.Status.MatchToOrderStatus())
		if err != nil {
			logger.Logger.Errorf("Updater failed to update order status: %v", err)
		}
		return
	}
	err = u.repo.UpdateOrderAccrual(ctx, order.Number, order.UserID, info.Status.MatchToOrderStatus(), info.Accrual)
	if err != nil {
		logger.Logger.Errorf("Updater failed to update order accrual: %v", err)
	}

}

func (u *Updater) waitUntilNextRequestAllowed(ctx context.Context) bool {
	for {
		u.mu.RLock()
		next := u.waitRequest
		u.mu.RUnlock()

		if next.IsZero() || time.Now().After(next) {
			return false
		}
		waitDuration := time.Until(next)
		select {
		case <-ctx.Done():
			return true
		case <-time.After(waitDuration):
		}
	}
}

func (u *Updater) updateNextRequestAllowed(timeOut int) {
	waitUntil := time.Now().Add(time.Duration(timeOut) * time.Second)
	u.mu.Lock()
	defer u.mu.Unlock()
	if waitUntil.After(u.waitRequest) {
		u.waitRequest = waitUntil
	}
}
