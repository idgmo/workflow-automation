package browser

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
)

// ExecuteLegacyExport launches an automated browser window to download files from legacy hubs
func ExecuteLegacyExport(loginURL, user, pass, targetDownloadSelector string) (string, error) {
	// Initialize playwright runtime core engine
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("could not start playwright engine: %w", err)
	}
	defer pw.Stop()

	// Launch a headless Chromium browser instance
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("could not launch automated browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return "", err
	}

	// Navigate to the target legacy portal URL
	if _, err = page.Goto(loginURL); err != nil {
		return "", fmt.Errorf("navigation failed: %w", err)
	}

	// Perform human action simulations: Enter security credentials and submit
	if err = page.Locator("#username").Fill(user); err != nil {
		return "", err
	}
	if err = page.Locator("#password").Fill(pass); err != nil {
		return "", err
	}
	if err = page.Locator("#login-btn").Click(); err != nil {
		return "", err
	}

	// Wait for the secure internal dashboard view to load, then trigger the data export click
	download, err := page.ExpectDownload(func() error {
		return page.Locator(targetDownloadSelector).Click()
	})
	if err != nil {
		return "", fmt.Errorf("data export trigger failed: %w", err)
	}

	// Save the output spreadsheet locally to pass to your pkg/spreadsheet engine
	localSavePath := "./downloads/daily_export.xlsx"
	if err = download.SaveAs(localSavePath); err != nil {
		return "", err
	}

	return localSavePath, nil
}
