package fixer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.sr.ht/~atmosx/calais/pkg/log"
	"git.sr.ht/~atmosx/calais/pkg/providers"
)

const apiBaseURL = "http://data.fixer.io/api/latest"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	apiKey string
	client HTTPDoer
	logger *log.Logger
}

type response struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	Error     struct {
		Info string `json:"info"`
	} `json:"error"`
}

func New(apiKey string, client HTTPDoer, logger *log.Logger) *Client {
	return &Client{apiKey: apiKey, client: client, logger: logger}
}

func (c *Client) FetchCurrency(from, to string) (*providers.CurrencyData, error) {
	url := fmt.Sprintf("%s?access_key=%s&base=%s&symbols=%s", apiBaseURL, c.apiKey, from, to)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if !r.Success {
		return nil, fmt.Errorf("fixer: %s", r.Error.Info)
	}

	rate, ok := r.Rates[to]
	if !ok {
		return nil, fmt.Errorf("unknown currency pair %s/%s", from, to)
	}

	return &providers.CurrencyData{
		From: from,
		To:   to,
		Rate: rate,
		Date: time.Unix(r.Timestamp, 0),
	}, nil
}
