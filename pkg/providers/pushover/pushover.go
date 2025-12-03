package pushover

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	pushoverAPIURL = "https://api.pushover.net/1/messages.json"
)

type Notifier interface {
	Send(title, message string) error
}

type Pushover struct {
	token     string
	recipient string
	client    *http.Client
}

func New(token, recipient string, client *http.Client) *Pushover {
	return &Pushover{
		token:     token,
		recipient: recipient,
		client:    client,
	}
}

func (p *Pushover) Send(title, message string) error {
	if p.token == "" || p.recipient == "" {
		return fmt.Errorf("pushover token or recipient is missing")
	}

	data := url.Values{}
	data.Set("token", p.token)
	data.Set("user", p.recipient)
	data.Set("title", title)
	data.Set("message", message)

	resp, err := p.client.PostForm(pushoverAPIURL, data)
	if err != nil {
		return fmt.Errorf("failed to send pushover request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover api returned non-200 status: %s", resp.Status)
	}

	return nil
}
