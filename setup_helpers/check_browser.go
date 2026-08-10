package setup_helpers

import (
	"fmt"
	"os"

	"github.com/mxschmitt/playwright-go"
)

func CheckBrowser() {
	pw, err := playwright.Run()
	if err != nil {
		fmt.Printf("Could not start playwright: %v\n", err)
		os.Exit(1)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		fmt.Printf("Could not launch browser (missing deps?): %v\n", err)
		os.Exit(1)
	}
	defer browser.Close()

	fmt.Println("Playwright Chromium browser launched successfully!")
}
