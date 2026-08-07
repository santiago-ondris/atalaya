package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 1 << 20

type Error struct {
	StatusCode int
	Retryable  bool
	Message    string
}

func (err *Error) Error() string { return err.Message }

type Client struct {
	endpoint   string
	chatID     string
	httpClient *http.Client
}

func NewClient(token, chatID string, timeout time.Duration, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{endpoint: "https://api.telegram.org/bot" + token + "/sendMessage", chatID: chatID, httpClient: httpClient}
}

func (client *Client) Send(ctx context.Context, text string) (int64, int, error) {
	payload, err := json.Marshal(map[string]any{
		"chat_id": client.chatID, "text": text, "parse_mode": "HTML", "disable_web_page_preview": true,
	})
	if err != nil {
		return 0, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint, bytes.NewReader(payload))
	if err != nil {
		return 0, 0, fmt.Errorf("build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(req)
	if err != nil {
		return 0, 0, &Error{Retryable: true, Message: "request Telegram: " + err.Error()}
	}
	defer response.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if readErr != nil {
		return 0, response.StatusCode, &Error{StatusCode: response.StatusCode, Retryable: true, Message: "read Telegram response: " + readErr.Error()}
	}
	var decoded struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return 0, response.StatusCode, &Error{StatusCode: response.StatusCode, Retryable: response.StatusCode >= 500, Message: "decode Telegram response"}
	}
	if response.StatusCode != http.StatusOK || !decoded.OK {
		retryable := response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		message := strings.TrimSpace(decoded.Description)
		if message == "" {
			message = "Telegram returned HTTP " + strconv.Itoa(response.StatusCode)
		}
		return 0, response.StatusCode, &Error{StatusCode: response.StatusCode, Retryable: retryable, Message: message}
	}
	return decoded.Result.MessageID, response.StatusCode, nil
}
