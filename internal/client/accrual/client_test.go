package accrual

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_GetOrderInfo(t *testing.T) {
	tests := []struct {
		name      string
		getClient func() (*httptest.Server, *HTTPClient)
		want      *OrderInfo
		expErr    func(*testing.T, error)
	}{
		{
			name: "url join error",
			getClient: func() (*httptest.Server, *HTTPClient) {
				return nil, NewHTTPClient("://invalid", 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GetOrderInfo join URL error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "server closed",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				s.Close()
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GetOrderInfo make request error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "OrderNotRegisteredError",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Equal(t, OrderNotRegisteredError, err)
			},
		},
		{
			name: "rate limit correct",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Retry-After", "30")
					w.WriteHeader(http.StatusTooManyRequests)
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				r := &RateLimitError{30, nil}
				assert.Equal(t, r, err)
			},
		},
		{
			name: "rate limit incorrect",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTooManyRequests)
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				r := &RateLimitError{0, nil}
				assert.Equal(t, r, err)
			},
		},
		{
			name: "internal error",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GetOrderInfo unexpected status code:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "marshall error",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GetOrderInfo unmarshal response error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Ok",
			getClient: func() (*httptest.Server, *HTTPClient) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					body := `{"order": "12345", "status": "PROCESSED", "accrual": 25.55}`
					w.Write([]byte(body))
				}))
				return s, NewHTTPClient(s.URL, 10)
			},
			want: &OrderInfo{"12345", "PROCESSED", 25.55},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GetOrderInfo unmarshal response error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, c := tt.getClient()
			if s != nil {
				defer s.Close()
			}
			got, err := c.GetOrderInfo("12345")
			if err != nil {
				assert.Error(t, err, tt.name)
				tt.expErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, "GetOrderInfo(%v) unexpected response", tt.name)
		})
	}
}
