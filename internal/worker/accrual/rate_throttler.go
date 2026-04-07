package accrual

import (
	"context"
	"sync"
	"time"
)

// RateThrottler реализует логику ожидания при превышении частоты запросов к стороннему сервису
type RateThrottler struct {
	mu        sync.RWMutex
	waitUntil time.Time
}

func newRateThrottler() *RateThrottler {
	return &RateThrottler{sync.RWMutex{}, time.Time{}}
}

// WaitUntilNextRequestAllowed логика проверки возможности и ожидания перед запросом
func (r *RateThrottler) WaitUntilNextRequestAllowed(ctx context.Context) bool {
	for {
		r.mu.RLock()
		next := r.waitUntil
		r.mu.RUnlock()

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

// UpdateNextRequestAllowed обновление переменной, содержащей время, ранее которого нельзя совершать запросы
func (r *RateThrottler) UpdateNextRequestAllowed(timeOut int) {
	waitUntil := time.Now().Add(time.Duration(timeOut) * time.Second)
	r.mu.Lock()
	defer r.mu.Unlock()
	if waitUntil.After(r.waitUntil) {
		r.waitUntil = waitUntil
	}
}
