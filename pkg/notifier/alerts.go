package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordPayload defines the required API JSON body structure
type DiscordPayload struct {
	Content string `json:"content"`
}

// Client manages transmission for outward alert channels - discord, text, etc.
type Client struct {
	WebhookURL string
	HTTPClient *http.Client
}

// NewClient initializes the notifier block with a secure 5-second connection timeout
func NewClient(webhookURL string) *Client {
	return &Client{
		WebhookURL: webhookURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// SendAlert fires a high-visibility text notification directly to a device - here it is the Discord app
func (c *Client) SendAlert(ctx context.Context, clientName, errorMessage string) error {
	if c.WebhookURL == "" {
		return fmt.Errorf("notifier webhook url is empty; alert dropped")
	}

	// Format a scannable message block for a mobile screen
	formattedText := fmt.Sprintf(
		"🚨 [B2B ALARM - %s]\n**Client:** %s\n**Error:** %s\n**Timestamp:** %s",
		time.Now().Format("15:04:05 MST"),
		clientName,
		errorMessage,
		time.Now().Format(time.RFC1123),
	)

	payload := DiscordPayload{Content: formattedText}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to build alert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to transmit alert over network: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns HTTP 204 No Content upon a successful webhook post
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord API rejected alert payload with status: %d", resp.StatusCode)
	}

	return nil
}
