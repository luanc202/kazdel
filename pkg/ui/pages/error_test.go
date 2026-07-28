package pages

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestErrorPage(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		title       string
		message     string
		expectMatch func(*testing.T, *goquery.Document)
	}{
		{
			name:    "renders 404 page",
			code:    http.StatusNotFound,
			title:   "Not Found",
			message: "The requested resource could not be found.",
			expectMatch: func(t *testing.T, doc *goquery.Document) {
				h1 := doc.Find("h1").Text()
				if h1 != "404" {
					t.Errorf("expected 404 in h1, got %q", h1)
				}

				h2 := doc.Find("h2").Text()
				if h2 != "Not Found" {
					t.Errorf("expected 'Not Found' in h2, got %q", h2)
				}

				p := doc.Find("p").Text()
				if p != "The requested resource could not be found." {
					t.Errorf("expected paragraph message, got %q", p)
				}
			},
		},
		{
			name:    "renders 500 page",
			code:    http.StatusInternalServerError,
			title:   "Internal Error",
			message: "Something went wrong.",
			expectMatch: func(t *testing.T, doc *goquery.Document) {
				h1 := doc.Find("h1").Text()
				if h1 != "500" {
					t.Errorf("expected 500 in h1, got %q", h1)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create a buffer to capture the templ component output
			var buf bytes.Buffer

			// Act: Render the component
			err := ErrorPage(tt.code, tt.title, tt.message).Render(context.Background(), &buf)
			if err != nil {
				t.Fatalf("failed to render component: %v", err)
			}

			// Assert: Parse HTML using goquery
			doc, err := goquery.NewDocumentFromReader(&buf)
			if err != nil {
				t.Fatalf("failed to parse HTML with goquery: %v", err)
			}

			// Run specific assertions
			tt.expectMatch(t, doc)
		})
	}
}
