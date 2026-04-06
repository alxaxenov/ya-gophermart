package accrual

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/go-resty/resty/v2"
)

var OrderNotRegisteredError = errors.New("order not registered")

type RateLimitError struct {
	Timeout int
	Err     error
}

func (r *RateLimitError) Error() string {
	return fmt.Sprintf("ratelimit timeout %d", r.Timeout)
}

func (r *RateLimitError) Unwrap() error {
	return r.Err
}

func newRateLimitError(req *resty.Response) *RateLimitError {
	timeout := req.Header().Get("Retry-After")
	seconds, err := strconv.Atoi(timeout)
	if err != nil {
		logger.Logger.Error("newRateLimitError parsing to int error", "error", err)
		seconds = 0
	}
	return &RateLimitError{Timeout: seconds, Err: nil}
}
