package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	fontsDir  = "pkg/ui/static/fonts"
	cssFile   = "pkg/ui/static/fonts.css"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
)

// FontURLs contains the Google Fonts CSS URLs you want to self-host.
// To add a new font, simply add its Google Fonts URL to this list and run the script!
var fontURLs = []string{
	"https://fonts.googleapis.com/css2?family=Libre+Franklin:ital,wght@0,100..900;1,100..900&display=swap",
	"https://fonts.googleapis.com/css2?family=Orbitron:wght@400..900&display=swap",
	"https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
}

func main() {
	if err := os.MkdirAll(fontsDir, 0755); err != nil {
		fmt.Printf("Error creating fonts directory: %v\n", err)
		return
	}

	var combinedCSS strings.Builder
	// Regex to extract .woff2 URLs from the Google Fonts CSS response
	urlRegex := regexp.MustCompile(`url\((https://[^)]+\.woff2)\)`)

	for _, fontURL := range fontURLs {
		fmt.Printf("Fetching CSS from: %s\n", fontURL)
		cssContent, err := fetchURL(fontURL)
		if err != nil {
			fmt.Printf("Error fetching font CSS: %v\n", err)
			continue
		}

		matches := urlRegex.FindAllStringSubmatch(cssContent, -1)
		newCSS := cssContent

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			fullMatch := match[0]
			fileURL := match[1]

			fileName := filepath.Base(fileURL)
			localPath := filepath.Join(fontsDir, fileName)

			// Download the font if it doesn't already exist locally
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				fmt.Printf("Downloading %s...\n", fileName)
				if err := downloadFile(fileURL, localPath); err != nil {
					fmt.Printf("Error downloading file %s: %v\n", fileURL, err)
					continue
				}
			}

			// Replace the remote URL with the local relative path inside the CSS
			localCSSPath := fmt.Sprintf("url(/static/fonts/%s)", fileName)
			newCSS = strings.Replace(newCSS, fullMatch, localCSSPath, -1)
		}

		combinedCSS.WriteString(fmt.Sprintf("/* Imported from: %s */\n", fontURL))
		combinedCSS.WriteString(newCSS)
		combinedCSS.WriteString("\n\n")
	}

	// Overwrite the fonts.css file to make the script idempotent
	if err := os.WriteFile(cssFile, []byte(combinedCSS.String()), 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", cssFile, err)
		return
	}

	fmt.Printf("Successfully processed all fonts! Saved to %s\n", cssFile)
}

func fetchURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	// We use an explicit User-Agent so Google sends us the modern woff2 CSS content
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func downloadFile(url string, filepath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
