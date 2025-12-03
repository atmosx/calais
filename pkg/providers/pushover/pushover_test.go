package pushover

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSend(t *testing.T) {
	// 1. Mock the Pushover API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/1/messages.json" {
			t.Errorf("Expected path /1/messages.json, got %s", r.URL.Path)
		}

		// Verify that the necessary form values are sent
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("token") != "test-token" {
			t.Errorf("Expected token 'test-token', got %s", r.FormValue("token"))
		}
		if r.FormValue("user") != "test-recipient" {
			t.Errorf("Expected user 'test-recipient', got %s", r.FormValue("user"))
		}
		if r.FormValue("message") != "Hello World" {
			t.Errorf("Expected message 'Hello World', got %s", r.FormValue("message"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Note: In a real scenario, we cannot easily change the const pushoverAPIURL.
	// However, for unit testing structs that accept a base URL or using a custom Transport
	// is usually preferred. Since the requested package structure is simple and the URL is const,
	// we will intercept the request using a custom Transport in the client to route
	// traffic to our mock server.

	client := &http.Client{
		Transport: &testTransport{
			Transport: http.DefaultTransport,
			URL:       mockServer.URL,
		},
	}

	// 2. Initialize the provider
	p := New("test-token", "test-recipient", client)

	// 3. Test sending a notification
	err := p.Send("Test Title", "Hello World")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSend_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Simulate API error
	}))
	defer mockServer.Close()

	client := &http.Client{
		Transport: &testTransport{
			Transport: http.DefaultTransport,
			URL:       mockServer.URL,
		},
	}

	p := New("test-token", "test-recipient", client)

	err := p.Send("Test", "Fail")
	if err == nil {
		t.Fatal("Expected error due to non-200 status, got nil")
	}
}

// testTransport intercepts requests and rewrites the scheme and host to the mock server
type testTransport struct {
	Transport http.RoundTripper
	URL       string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, _ := url.Parse(t.URL)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	return t.Transport.RoundTrip(req)
}
