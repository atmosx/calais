package fixer

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"git.sr.ht/~atmosx/calais/pkg/log"
	"git.sr.ht/~atmosx/calais/pkg/providers"
)

type mockHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func newTestClient(fn func(req *http.Request) (*http.Response, error)) *Client {
	logger := log.New(io.Discard, "Error")
	return New("test-key", &mockHTTPClient{do: fn}, logger)
}

func TestFetchCurrency(t *testing.T) {
	tests := []struct {
		name        string
		mockResp    string
		mockStatus  int
		mockErr     error
		expectErr   bool
		checkResult func(*testing.T, *providers.CurrencyData)
	}{
		{
			name:       "success",
			mockStatus: http.StatusOK,
			mockResp: `{
                "success": true,
                "timestamp": 1666108800,
                "base": "EUR",
                "rates": {"USD": 1.05}
            }`,
			expectErr: false,
			checkResult: func(t *testing.T, cd *providers.CurrencyData) {
				if cd == nil {
					t.Fatal("expected result, got nil")
				}
				if cd.From != "EUR" || cd.To != "USD" || cd.Rate != 1.05 {
					t.Errorf("unexpected result: %+v", cd)
				}
				if cd.Date.Unix() != 1666108800 {
					t.Errorf("unexpected date: %v", cd.Date)
				}
			},
		},
		{
			name:       "api error",
			mockStatus: http.StatusOK,
			mockResp: `{
                "success": false,
                "error": {"info": "Invalid API key"}
            }`,
			expectErr: true,
		},
		{
			name:       "malformed json",
			mockStatus: http.StatusOK,
			mockResp:   `{"success": true, "rates": {"USD": 1.05`, // truncated
			expectErr:  true,
		},
		{
			name:       "unknown currency",
			mockStatus: http.StatusOK,
			mockResp: `{
                "success": true,
                "timestamp": 1666108800,
                "base": "EUR",
                "rates": {}
            }`,
			expectErr: true,
		},
		{
			name:      "network error",
			mockErr:   errors.New("network down"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(func(req *http.Request) (*http.Response, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return &http.Response{
					StatusCode: tt.mockStatus,
					Body:       io.NopCloser(bytes.NewBufferString(tt.mockResp)),
				}, nil
			})

			cd, err := client.FetchCurrency("EUR", "USD")

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, cd)
			}
		})
	}
}

func TestNew(t *testing.T) {
	logger := log.New(io.Discard, "Error")
	client := New("secret", &http.Client{}, logger)
	if client.apiKey != "secret" {
		t.Errorf("expected apiKey secret, got %s", client.apiKey)
	}
}
