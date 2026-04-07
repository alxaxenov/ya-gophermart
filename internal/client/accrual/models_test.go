package accrual

import (
	"testing"

	"github.com/alxaxenov/ya-gophermart/internal/domain/order"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name string
		e    OrderStatus
		want bool
	}{
		{
			name: "REGISTERED",
			e:    REGISTERED,
			want: false,
		},
		{
			name: "INVALID",
			e:    INVALID,
			want: true,
		},
		{
			name: "PROCESSING",
			e:    PROCESSING,
			want: false,
		},
		{
			name: "PROCESSED",
			e:    PROCESSED,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.e.IsFinal(), "IsFinal()")
		})
	}
}

func TestOrderStatus_MatchToOrderStatus(t *testing.T) {
	tests := []struct {
		name string
		e    OrderStatus
		want order.Status
	}{
		{
			name: "REGISTERED",
			e:    REGISTERED,
			want: order.PROCESSING,
		},
		{
			name: "INVALID",
			e:    INVALID,
			want: order.INVALID,
		},
		{
			name: "PROCESSING",
			e:    PROCESSING,
			want: order.PROCESSING,
		},
		{
			name: "PROCESSED",
			e:    PROCESSED,
			want: order.PROCESSED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.e.MatchToOrderStatus(), "MatchToOrderStatus()")
		})
	}
}
