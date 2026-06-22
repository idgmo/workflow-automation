package stripeClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ChargeRequest holds the uniform data payload required to execute a B2B transaction
type ChargeRequest struct {
	CustomerEmail string
	AmountCents   int64
	Currency      string
	Description   string
}

// Client acts as the reusable object engine for managing Stripe API traffic
type Client struct {
	SecretKey  string
	HTTPClient *http.Client
}

// NewClient instantiates the package with secure keys and a default 10-second request timeout
func NewClient(secretKey string) *Client {
	return &Client{
		SecretKey:  secretKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ExecuteCharge Pushes the transaction payload to Stripe, featuring an automated self-healing retry loop
func (c *Client) ExecuteCharge(ctx context.Context, charge ChargeRequest) (string, error) {
	// Stripe expects parameters in x-www-form-urlencoded format
	formData := url.Values{}
	formData.Set("amount", fmt.Sprintf("%d", charge.AmountCents))
	formData.Set("currency", strings.ToLower(charge.Currency))
	formData.Set("description", charge.Description)
	// formData.Set("receipt_email", charge.CustomerEmail)

	formData.Set("source", "tok_visa")

	endpoint := "https://api.stripe.com/v1/charges"

	// Self-healing retry settings: 2s -> 4s -> 8s -> 16s...
	baseDelay := 2 * time.Second
	maxDelay := 30 * time.Second
	maxAttempts := 4

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(formData.Encode()))
		if err != nil {
			return "", fmt.Errorf("failed to generate http request: %w", err)
		}

		// Formatting Cleanup
		c.SecretKey = strings.TrimSpace(c.SecretKey)
		c.SecretKey = strings.Trim(c.SecretKey, "\"")
		c.SecretKey = strings.Trim(c.SecretKey, "'")

		// Set essential authorization headers
		// req.Header.Set("Authorization", "Bearer "+c.SecretKey)
		req.SetBasicAuth(c.SecretKey, "")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			// Network-level failure occurred. Trigger backoff logic.
			if attempt == maxAttempts {
				return "", fmt.Errorf("network failure after %d attempts: %w", maxAttempts, err)
			}
			c.waitBackoff(attempt, baseDelay, maxDelay)
			continue
		}

		// Parse the API JSON response safely
		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		resp.Body.Close() // Immediately close stream to prevent severe memory leaks

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			// Transaction completely successful! Extract the unique transaction ID
			if id, ok := responseBody["id"].(string); ok {
				return id, nil
			}
			return "SUCCESS_NO_ID", nil
		}

		// Handle explicit API Denials/Errors from Stripe (e.g., card declined, invalid key)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			if apiErr, ok := responseBody["error"].(map[string]interface{}); ok {
				return "", fmt.Errorf("stripe api rejected request (Status %d): %v", resp.StatusCode, apiErr["message"])
			}
			return "", fmt.Errorf("stripe api rejected request with status code: %d", resp.StatusCode)
		}

		// If it's a 500-level server error from Stripe's end, pause and retry
		if attempt == maxAttempts {
			return "", fmt.Errorf("stripe server error sustained over %d attempts (Status %d)", maxAttempts, resp.StatusCode)
		}
		c.waitBackoff(attempt, baseDelay, maxDelay)
	}

	return "", fmt.Errorf("transaction failed to complete within the maximum execution limit")
}

// waitBackoff calculates and executes the pause duration using an exponential math shift
func (c *Client) waitBackoff(attempt int, base, max time.Duration) {
	delay := base * (1 << uint(attempt-1))
	if delay > max {
		delay = max
	}
	time.Sleep(delay)
}
