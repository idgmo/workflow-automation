package imapClient

import (
	"log"
	"time"
)

// SecureConnect attempts to link to the server, retrying with increasing delays if it fails
func SecureConnect(server string) (*ClientInstance, error) {
	baseDelay := 2 * time.Second
	maxDelay := 5 * time.Minute
	attempts := 0

	for {
		client, err := dialTLS(server)
		if err == nil {
			return client, nil // Connection successful!
		}

		attempts++
		log.Printf("[RETRY SYSTEM] Connection failed (Attempt %d): %v", attempts, err)

		// Calculate the next delay: 2s -> 4s -> 8s -> 16s... up to 5 minutes max
		currentDelay := baseDelay * (1 << uint(attempts))
		if currentDelay > maxDelay {
			currentDelay = maxDelay
		}

		log.Printf("[RETRY SYSTEM] Standing by for %v before attempting next connection...", currentDelay)
		time.Sleep(currentDelay)
	}
}
