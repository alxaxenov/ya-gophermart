package client

import (
	"testing"
)

func TestUrlJoin(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		paths    []string
		wantURL  string
		wantErr  bool
	}{
		{
			name:     "One path",
			basePath: "http://example.com",
			paths:    []string{"api"},
			wantURL:  "http://example.com/api",
			wantErr:  false,
		},
		{
			name:     "Several paths",
			basePath: "http://example.com",
			paths:    []string{"api", "v1", "users"},
			wantURL:  "http://example.com/api/v1/users",
			wantErr:  false,
		},
		{
			name:     "BaseURL with path",
			basePath: "http://example.com/api/",
			paths:    []string{"v1", "users"},
			wantURL:  "http://example.com/api/v1/users",
			wantErr:  false,
		},
		{
			name:     "Incorrect baseUrl",
			basePath: "http://%zz",
			paths:    []string{"path"},
			wantURL:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UrlJoin(tt.basePath, tt.paths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("UrlJoin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Errorf("UrlJoin() вернул nil, хотя ошибки не ожидалось")
				return
			}
			if got.String() != tt.wantURL {
				t.Errorf("UrlJoin() = %v, want %v", got.String(), tt.wantURL)
			}
		})
	}
}
