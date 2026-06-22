package main

import (
	"context"
	"log"
	"os"

	"workflowAutomation/pkg/emailParser"
	"workflowAutomation/pkg/notifier"
	"workflowAutomation/pkg/stripeClient"
)

func main() {
	ctx := context.Background()
	clientName := "Alpine Properties LLC"

	// Gather pipeline configurations from safe server environment strings
	discordURL := os.Getenv("DISCORD_WEBHOOK_URL")
	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")

	// Instantiate blocks
	alertEngine := notifier.NewClient(discordURL)
	stripeGateway := stripeClient.NewClient(stripeSecret)

	// Simulated incoming malformed message missing critical unit location parameters
	// mockCorruptedEmail := "Hello, someone left an invoice statement on my desk. Charge the client card $50."
	mockCorrectEmail := "Hello, I have finished the plumbing issue in Apt 4B. Please charge $50 when available."

	log.Printf("[%s] Incoming task received. Processing data stack...", clientName)

	// 1. Process Phase
	parsedData, err := emailParser.ParseEmailBody(mockCorrectEmail)
	if err != nil {
		// Log internally and fire out an alert to your phone via the notifier package
		log.Printf("[CRITICAL] Internal text parse failed: %v", err)
		_ = alertEngine.SendAlert(ctx, clientName, "Email processing step failed: "+err.Error())
		return
	}

	// 2. Fallback Safety Filter: If our parser outputs a default "UNKNOWN" unit number flag
	if parsedData.UnitNumber == "UNKNOWN" {
		errMessage := "Extraction failed: Script could not isolate a valid apartment unit code from message text"
		log.Printf("[CRITICAL] %s", errMessage)

		// Ping your mobile Discord app instantly
		_ = alertEngine.SendAlert(ctx, clientName, errMessage)
		return
	}

	// 3. Output Phase (Only runs if data clears the safety filter rules above)
	payload := stripeClient.ChargeRequest{
		CustomerEmail: "tenant_fallback@example.com",
		AmountCents:   5000,
		Currency:      "usd",
		Description:   "Automated rental processing fee",
	}

	_, err = stripeGateway.ExecuteCharge(ctx, payload)
	if err != nil {
		log.Printf("[CRITICAL] Financial gateway rejected payload: %v", err)
		_ = alertEngine.SendAlert(ctx, clientName, "Stripe transmission failure: "+err.Error())
		return
	}

	log.Println("[SUCCESS] End-to-end pipeline finished execution without errors.")
}
