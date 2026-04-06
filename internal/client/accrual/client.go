package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/client"
	"github.com/go-resty/resty/v2"
)

type HTTPClient struct {
	baseURL string
	client  *resty.Client
}

func NewHTTPClient(baseURL string, timeout int) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  resty.New().SetTimeout(time.Duration(timeout) * time.Second),
	}
}

// GetOrderInfo получение информации о вознаграждении за заказ
func (c *HTTPClient) GetOrderInfo(orderID string) (*OrderInfo, error) {
	const (
		endpoint = "/api/orders/%s"
		method   = resty.MethodGet
	)

	reqURL, err := client.UrlJoin(c.baseURL, fmt.Sprintf(endpoint, orderID))
	if err != nil {
		return nil, fmt.Errorf("GetOrderInfo join URL error: %w", err)
	}

	resp, err := c.client.R().Execute(method, reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("GetOrderInfo make request error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNoContent:
			return nil, OrderNotRegisteredError
		case http.StatusTooManyRequests:
			return nil, newRateLimitError(resp)
		default:
			return nil, fmt.Errorf("GetOrderInfo unexpected status code: %d, data: %s", resp.StatusCode(), resp.String())
		}
	}

	var result OrderInfo
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("GetOrderInfo unmarshal response error: %w", err)
	}

	return &result, nil
}
