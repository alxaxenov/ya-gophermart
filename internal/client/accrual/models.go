package accrual

import "github.com/alxaxenov/ya-gophermart/internal/domain/order"

type OrderStatus string

const (
	REGISTERED OrderStatus = "REGISTERED"
	INVALID    OrderStatus = "INVALID"
	PROCESSING OrderStatus = "PROCESSING"
	PROCESSED  OrderStatus = "PROCESSED"
)

func (e OrderStatus) IsFinal() bool {
	return e == INVALID || e == PROCESSED
}

func (e OrderStatus) MatchToOrderStatus() order.Status {
	switch e {
	case REGISTERED, PROCESSING:
		return order.PROCESSING
	case PROCESSED:
		return order.PROCESSED
	case INVALID:
		return order.INVALID
	default:
		return order.PROCESSING
	}
}

type OrderInfo struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float64     `json:"accrual"`
}
