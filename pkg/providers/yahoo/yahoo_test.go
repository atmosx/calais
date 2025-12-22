package yahoo

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/atmosx/calais/pkg/log"
	"github.com/atmosx/calais/pkg/providers"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func newTestClient(fn func(req *http.Request) (*http.Response, error)) *Client {
	logger := log.New(io.Discard, "Error")
	return New(&mockHTTPClient{DoFunc: fn}, logger)
}

func TestFetchStock(t *testing.T) {
	tests := []struct {
		name        string
		mockDoFunc  func(req *http.Request) (*http.Response, error)
		expectError bool
		check       func(t *testing.T, sd *providers.StockData, err error)
	}{
		{
			name: "success",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				jsonResp := `{
"chart": {
						"result": [{
							"meta": { "symbol": "AAPL" },
							"timestamp": [1755475200],
							"indicators": {
								"quote": [{
									"close": [150.75],
									"volume": [12345678]
								}]
							}
						}],
						"error": null
					}
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				}, nil
			},
			expectError: false,
			check: func(t *testing.T, sd *providers.StockData, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if sd.Symbol != "AAPL" || sd.Close != 150.75 {
					t.Errorf("unexpected data: %+v", sd)
				}
				if sd.Date.Year() != 2025 || sd.Date.Month() != time.August || sd.Date.Day() != 18 {
					t.Errorf("date mismatch: %v", sd.Date)
				}
			},
		},
		{
			name: "http error",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewBufferString("not found")),
				}, nil
			},
			expectError: true,
		},
		{
			name: "network error",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("simulated network failure")
			},
			expectError: true,
		},
		{
			name: "malformed JSON",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"chart": { "result": [`)),
				}, nil
			},
			expectError: true,
		},
		{
			name: "yahoo api error",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				jsonResp := `{
"chart": {
						"result": null,
						"error": {
							"code": "Not Found",
							"description": "Symbol not found"
						}
					}
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				}, nil
			},
			expectError: true,
		},
		{
			name: "empty data",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"chart":{"result":[]}}`)),
				}, nil
			},
			expectError: true,
		},
		{
			name: "empty timestamp/indicators",
			mockDoFunc: func(req *http.Request) (*http.Response, error) {
				jsonResp := `{
"chart": {
						"result": [{
							"meta": { "symbol": "AAPL" },
							"timestamp": [],
							"indicators": { "quote": [] }
						}],
						"error": null
					}
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.mockDoFunc)
			sd, err := client.FetchStock("TEST")
			switch {
			case tt.expectError && err == nil:
				t.Fatal("expected an error but got none")
			case !tt.expectError && err != nil:
				t.Fatalf("did not expect an error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, sd, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	logger := log.New(os.Stdout, "Info")
	httpClient := &http.Client{}

	client := New(httpClient, logger)

	if client.client != httpClient {
		t.Error("httpClient not wired correctly")
	}
	if client.logger != logger {
		t.Error("logger not wired correctly")
	}
}
