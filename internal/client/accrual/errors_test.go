package accrual

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func Test_newRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		retryValue string
		want       *RateLimitError
	}{
		{
			name:       "Ok",
			retryValue: "60",
			want:       &RateLimitError{60, nil},
		},
		{
			name:       "no header",
			retryValue: "",
			want:       &RateLimitError{0, nil},
		},
		{
			name:       "incorrect header value",
			retryValue: "abc",
			want:       &RateLimitError{0, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawResp := &http.Response{Header: make(http.Header)}
			if tt.retryValue != "" {
				rawResp.Header.Add("Retry-After", tt.retryValue)
			}
			resp := &resty.Response{RawResponse: rawResp}

			got := newRateLimitError(resp)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRateLimitError_Error(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		want    string
	}{
		{
			name:    "not zero",
			timeout: 10,
			want:    "ratelimit timeout 10",
		},
		{
			name:    "zero",
			timeout: 0,
			want:    "ratelimit timeout 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RateLimitError{
				Timeout: tt.timeout,
				Err:     nil,
			}
			assert.Equalf(t, tt.want, r.Error(), "Error()")
		})
	}
}
