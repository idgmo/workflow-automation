package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
	"workflowAutomation/pkg/emailParser"
	"workflowAutomation/pkg/localDatabase"
	"workflowAutomation/pkg/notifier"
	"workflowAutomation/pkg/stripeClient"
)

func executeTransactionWithTimeout(stripeGateway *stripeClient.Client, payload stripeClient.ChargeRequest) error {
	// Create an isolated context that self-destructs after precisely 15 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // Releases internal timer resources immediately once execution finishes

	// Pass this timeout context directly down into the network package block
	txID, err := stripeGateway.ExecuteCharge(ctx, payload)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("[CRITICAL TIMEOUT] Network dropped out permanently. Operation canceled safely.")
			// Trigger a notification alert to a phone here
			return err
		}
		log.Printf("[PIPELINE ERROR] Operational fault: %v\n", err)
		return err
	}

	log.Printf("[SUCCESS] Transaction processed: %s\n", txID)
	return nil
}

func main() {
	ctx := context.Background()

	// Gather pipeline configurations from server environment strings
	clientName := os.Getenv("CLIENT_NAME")
	discordURL := os.Getenv("DISCORD_WEBHOOK_URL")
	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")

	// Database bindings
	// Maps to the internal directory generated inside Dockerfile
	dbFolder := "/root/database"
	dbPath := filepath.Join(dbFolder, "local_ledger.db")

	// Build the internal folders safely if missing
	if err := os.MkdirAll(dbFolder, 0755); err != nil {
		log.Fatalf("[CRITICAL INIT] Failed to create internal storage folder: %v", err)
	}

	// Initialize database module
	store, err := localDatabase.NewStore(dbPath)
	if err != nil {
		log.Fatalf("[CRITICAL INIT] Database setup aborted: %v", err)
	}
	defer store.Close()

	log.Printf("[%s] Database initialized securely at %s", clientName, dbPath)

	// Instantiate blocks
	alertEngine := notifier.NewClient(discordURL)
	stripeGateway := stripeClient.NewClient(stripeSecret)

	// Simulated incoming malformed message missing required unit location parameters
	// mockCorruptEmail := "Hello, someone left an invoice statement on my desk. Charge the client card $50."

	mockCorrectEmail := "Hello, I have finished the plumbing issue in Apt 4B. Please charge $50 when available."

	log.Printf("[%s] Incoming task received. Processing data stack...", clientName)

	// 1. Process Phase
	parsedData, err := emailParser.ParseEmailBody(mockCorrectEmail)
	if err != nil {
		// Log internally and fire out an alert to a specific phone via the notifier package
		log.Printf("[CRITICAL] Internal text parse failed: %v", err)
		_ = alertEngine.SendAlert(ctx, clientName, "Email processing step failed: "+err.Error())
		return
	}

	// 2. Fallback Safety Filter: If our parser outputs a default "UNKNOWN" unit number flag
	if parsedData.UnitNumber == "UNKNOWN" {
		errMessage := "Extraction failed: Internal script could not isolate a valid apartment unit code from message text"
		log.Printf("[CRITICAL] %s", errMessage)

		// Ping the mobile Discord app instantly - as notifier
		_ = alertEngine.SendAlert(ctx, clientName, errMessage)
		return
	}

	// 3. Output Phase (Only runs if data clears the above safety filters)
	payload := stripeClient.ChargeRequest{
		CustomerEmail: "tenant_fallback@example.com",
		AmountCents:   5000,
		Currency:      "usd",
		Description:   "Automated rental processing fee",
	}

	err = executeTransactionWithTimeout(stripeGateway, payload)
	if err != nil {
		log.Printf("[CRITICAL] Financial gateway rejected payload: %v", err)
		// Notify through Discord
		_ = alertEngine.SendAlert(ctx, clientName, "Stripe transmission failure: "+err.Error())
		return
	}

	log.Println("[SUCCESS] End-to-end pipeline finished execution without errors.")
}
