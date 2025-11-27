package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.sr.ht/~atmosx/calais/pkg/log"
	"git.sr.ht/~atmosx/calais/pkg/providers"
)

const apiBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client HTTPDoer
	logger *log.Logger
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol string `json:"symbol"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close  []*float64 `json:"close"`
					Volume []*float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func New(client HTTPDoer, logger *log.Logger) *Client {
	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) FetchStock(symbol string) (*providers.StockData, error) {
	url := fmt.Sprintf("%s/%s?interval=1d&range=1d", apiBaseURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to create request for symbol %s: %w", symbol, err)
	}

	// Yahoo Finance often rejects requests without a User-Agent header.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Calais/1.0)")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to fetch data for symbol %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Received non-OK HTTP status", "status", resp.Status, "symbol", symbol)
		return nil, fmt.Errorf("bad response status for symbol %s: %s", symbol, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to read response for %s: %w", symbol, err)
	}

	var yahooData yahooChartResponse
	if err := json.Unmarshal(body, &yahooData); err != nil {
		c.logger.Error("Failed to unmarshal JSON response", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to decode response for %s: %w", symbol, err)
	}

	if yahooData.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo api error: %s", yahooData.Chart.Error.Description)
	}

	if len(yahooData.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for symbol %s", symbol)
	}

	result := yahooData.Chart.Result[0]
	if len(result.Timestamp) == 0 || len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("empty time series for symbol %s", symbol)
	}

	idx := len(result.Timestamp) - 1
	quote := result.Indicators.Quote[0]

	if len(quote.Close) <= idx || quote.Close[idx] == nil {
		return nil, fmt.Errorf("missing close price for symbol %s", symbol)
	}

	var vol float64
	if len(quote.Volume) > idx && quote.Volume[idx] != nil {
		vol = *quote.Volume[idx]
	}

	return &providers.StockData{
		Symbol: result.Meta.Symbol,
		Date:   time.Unix(result.Timestamp[idx], 0),
		Close:  *quote.Close[idx],
		Volume: vol,
	}, nil
}
