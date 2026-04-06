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
	wg            sync.WaitGroup
	ch            chan model.OrderToProcess
	workersNum    int
	throttler     throttler
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type gophermartRepo interface {
	GetOrdersToProcess(context.Context, int) (*[]model.OrderToProcess, error)
	UpdateOrderStatus(context.Context, string, d_o.Status) error
	UpdateOrderAccrual(context.Context, string, int, d_o.Status, float64) error
}

type accrualClient interface {
	GetOrderInfo(string) (*accrual.OrderInfo, error)
}

type throttler interface {
	WaitUntilNextRequestAllowed(context.Context) bool
	UpdateNextRequestAllowed(int)
}

func NewUpdater(ctx context.Context, repo gophermartRepo, accrualClient accrualClient, workersNum int) *Updater {
	u := &Updater{
		repo:          repo,
		accrualClient: accrualClient,
		wg:            sync.WaitGroup{},
		ch:            make(chan model.OrderToProcess, workersNum*2),
		workersNum:    workersNum,
		throttler:     newRateThrottler(),
	}
	for i := 0; i < workersNum; i++ {
		go u.worker(ctx)
	}
	return u
}

// Run запуск фонового процесса обновления accrual
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
				logger.Logger.Error("Updater failed to get orders from db", "error", err)
			}
			if len(*orders) == 0 {
				continue
			}
			for _, order := range *orders {
				u.ch <- order
				u.wg.Add(1)
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

// innerWorker основная логика обновления поля accrual
// Запрос в сторонний сервис, рассчитывающий вознаграждение, обновление полей accrual и status заказа
func (u *Updater) innerWorker(ctx context.Context, order *model.OrderToProcess) {
	defer u.wg.Done()

	if order.Status == d_o.NEW {
		err := u.repo.UpdateOrderStatus(ctx, order.Number, d_o.PROCESSING)
		if err != nil {
			logger.Logger.Error("Updater failed to update order initial status", "error", err)
			return
		}
	}

	ctx_done := u.throttler.WaitUntilNextRequestAllowed(ctx)
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
			logger.Logger.Info("Updater accrual timeout responce", "timeout", rateLimit.Timeout)
			u.throttler.UpdateNextRequestAllowed(rateLimit.Timeout)
			return
		default:
			logger.Logger.Error("Updater failed to get accrual order info", "error", err)
			return
		}
	}

	if !info.Status.IsFinal() {
		return
	}
	if info.Accrual == 0 {
		err = u.repo.UpdateOrderStatus(ctx, order.Number, info.Status.MatchToOrderStatus())
		if err != nil {
			logger.Logger.Error("Updater failed to update order status", "error", err)
		}
		return
	}
	err = u.repo.UpdateOrderAccrual(ctx, order.Number, order.UserID, info.Status.MatchToOrderStatus(), info.Accrual)
	if err != nil {
		logger.Logger.Error("Updater failed to update order accrual", "error", err)
	}
}
