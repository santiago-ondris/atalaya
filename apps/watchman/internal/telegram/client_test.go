package telegram

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

func TestSendSuccess(t *testing.T) {
	client := NewClient("secret-token", "123", time.Second, &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.Contains(request.URL.String(), "secret-token") == false {
			t.Fatal("token missing from Telegram endpoint")
		}
		body, _ := io.ReadAll(request.Body)
		if !strings.Contains(string(body), `"chat_id":"123"`) {
			t.Fatalf("unexpected request: %s", body)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":42}}`))}, nil
	})})
	messageID, status, err := client.Send(context.Background(), "hello")
	if err != nil || messageID != 42 || status != 200 {
		t.Fatalf("Send() = %d, %d, %v", messageID, status, err)
	}
}

func TestSendClassifiesFailures(t *testing.T) {
	tests := []struct {
		status    int
		retryable bool
	}{{429, true}, {503, true}, {401, false}, {400, false}}
	for _, test := range tests {
		client := NewClient("token", "123", time.Second, &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: test.status, Body: io.NopCloser(strings.NewReader(`{"ok":false,"description":"provider failure"}`))}, nil
		})})
		_, _, err := client.Send(context.Background(), "hello")
		providerErr, ok := err.(*Error)
		if !ok || providerErr.Retryable != test.retryable {
			t.Fatalf("status %d: error = %#v", test.status, err)
		}
	}
}
