package marketstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.sr.ht/~atmosx/calais/pkg/log"
	"git.sr.ht/~atmosx/calais/pkg/providers"
)

const apiBaseURL = "https://api.marketstack.com/v2/eod"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	apiKey string
	client HTTPDoer
	logger *log.Logger
}

type MarketstackTime time.Time

func (mt *MarketstackTime) UnmarshalJSON(b []byte) error {
	const layout = "2006-01-02T15:04:05-0700"
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return err
	}
	*mt = MarketstackTime(t)
	return nil
}

type marketstackResponse struct {
	Data []struct {
		Symbol string          `json:"symbol"`
		Date   MarketstackTime `json:"date"`
		Close  float64         `json:"close"`
		Volume float64         `json:"volume"`
	} `json:"data"`
}

func New(apiKey string, client HTTPDoer, logger *log.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		client: client,
		logger: logger,
	}
}

func (c *Client) FetchStock(symbol string) (*providers.StockData, error) {
	url := fmt.Sprintf("%s?access_key=%s&symbols=%s&latest=true&limit=1", apiBaseURL, c.apiKey, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to create request for symbol %s: %w", symbol, err)
	}

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

	var marketstackData marketstackResponse
	if err := json.Unmarshal(body, &marketstackData); err != nil {
		c.logger.Error("Failed to unmarshal JSON response", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to decode response for %s: %w", symbol, err)
	}

	if len(marketstackData.Data) == 0 {
		return nil, fmt.Errorf("no data returned for symbol %s", symbol)
	}

	stock := marketstackData.Data[0]
	return &providers.StockData{
		Symbol: stock.Symbol,
		Date:   time.Time(stock.Date),
		Close:  stock.Close,
		Volume: stock.Volume,
	}, nil
}
