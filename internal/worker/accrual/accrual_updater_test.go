package accrual

import (
	"context"
	"sync"
	"testing"

	"errors"

	"github.com/alxaxenov/ya-gophermart/internal/client/accrual"
	"github.com/alxaxenov/ya-gophermart/internal/domain/order"
	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/alxaxenov/ya-gophermart/internal/worker/accrual/mocks"
	"go.uber.org/mock/gomock"
)

func TestUpdater_innerWorker(t *testing.T) {
	tests := []struct {
		name      string
		order     *model.OrderToProcess
		mockSetup func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler)
	}{
		{
			name:  "new order UpdateOrderStatus error",
			order: &model.OrderToProcess{"123456", order.NEW, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), "123456", order.PROCESSING).Times(1).Return(errors.New("test error")),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(0)
				a.EXPECT().GetOrderInfo(gomock.Any()).Times(0)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "new order WaitUntilNextRequestAllowed context done",
			order: &model.OrderToProcess{"123456", order.NEW, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), "123456", order.PROCESSING).Times(1).Return(nil),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(true)
				a.EXPECT().GetOrderInfo(gomock.Any()).Times(0)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo OrderNotRegisteredError",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(nil, accrual.OrderNotRegisteredError)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo rateLimit response",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(nil, &accrual.RateLimitError{60, nil})
				th.EXPECT().UpdateNextRequestAllowed(60).Times(1)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo another error",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(nil, errors.New("test error"))
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo not final status",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(
					&accrual.OrderInfo{"123456", accrual.PROCESSING, 0}, nil,
				)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo invalid status",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), "123456", order.INVALID).Times(1),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(
					&accrual.OrderInfo{"123456", accrual.INVALID, 0}, nil,
				)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "processing order GetOrderInfo processed status",
			order: &model.OrderToProcess{"123456", order.PROCESSING, 123},
			mockSetup: func(r *mocks.MockgophermartRepo, a *mocks.MockaccrualClient, th *mocks.Mockthrottler) {
				gomock.InOrder(
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					r.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
				)
				th.EXPECT().WaitUntilNextRequestAllowed(gomock.Any()).Times(1).Return(false)
				a.EXPECT().GetOrderInfo("123456").Times(1).Return(
					&accrual.OrderInfo{"123456", accrual.PROCESSED, 25.55}, nil,
				)
				th.EXPECT().UpdateNextRequestAllowed(gomock.Any()).Times(0)
				r.EXPECT().UpdateOrderAccrual(gomock.Any(), "123456", 123, order.PROCESSED, 25.55).Times(1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockgophermartRepo(ctrl)
			client := mocks.NewMockaccrualClient(ctrl)
			th := mocks.NewMockthrottler(ctrl)
			tt.mockSetup(repo, client, th)

			u := &Updater{
				repo:          repo,
				accrualClient: client,
				wg:            sync.WaitGroup{},
				ch:            nil,
				workersNum:    0,
				throttler:     th,
			}
			u.wg.Add(1)
			u.innerWorker(context.Background(), tt.order)
		})
	}
}
