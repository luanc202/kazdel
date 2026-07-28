package e2e_test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kazdel/pkg/handlers"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/infra/db"
	"kazdel/pkg/infra/mail"
	"kazdel/pkg/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pw          *playwright.Playwright
	browser     playwright.Browser
	ts          *httptest.Server
	dbContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	// 0. Change working directory to project root so static files and migrations can load
	if err := os.Chdir("../../"); err != nil {
		log.Fatalf("failed to chdir to root: %v", err)
	}
	ctx := context.Background()

	// 1. Initialize Postgres Database container
	fmt.Println("Starting PostgreSQL test container...")
	var err error
	dbContainer, err = postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	connStr, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	// 2. Setup env vars
	os.Setenv("DATABASE_URL", connStr)
	os.Setenv("MIGRATION_DATABASE_URL", connStr)
	os.Setenv("PORT", "8080")
	os.Setenv("ENVIRONMENT", "testing")
	os.Setenv("MIGRATIONS_PATH", "migrations")
	os.Setenv("JWT_SECRET", "supersekret")
	os.Setenv("MAIL_ENABLED", "false")

	// 3. Initialize app configs
	env, err := config.LoadEnv(".") // Now at project root
	if err != nil {
		fmt.Printf("Note: %v\n", err)
	}

	err = config.InitConfigs()
	if err != nil {
		log.Fatalf("could not initialize configs: %v", err)
	}

	// 4. Run migrations
	mi, err := migrate.New("file://migrations/"+env.GetDatabaseType(), connStr)
	if err != nil {
		log.Fatalf("failed to initialize migration: %v", err)
	}
	if err := mi.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run up migrations: %v", err)
	}

	// 5. Initialize application
	dbConn := config.GetDbConnection()
	shortenedURLsRepo := db.NewShortenedUrlRepository(dbConn)
	userRepo := db.NewUserRepository(dbConn)
	sessionRepo := db.NewSessionRepository(dbConn)

	userTokenRepo := db.NewUserTokenRepository(dbConn)
	emailService := mail.NewSMTPMailService()

	urlVisitRepo := db.NewUrlVisitRepository(dbConn)
	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo, urlVisitRepo, nil, emailService)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, userTokenRepo, emailService)

	deps := &handlers.Dependencies{
		ShortenedUrlUseCase: shortenedURLUseCase,
		AuthUseCase:         authUseCase,
	}

	router, err := handlers.BuildRouter(deps)
	if err != nil {
		log.Fatalf("failed to build router: %v", err)
	}

	// 6. Start the test server
	ts = httptest.NewServer(router)
	fmt.Printf("Test server started at %s\n", ts.URL)

	// 7. Initialize Playwright
	err = playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	// Run all tests
	code := m.Run()

	// Teardown
	browser.Close()
	pw.Stop()
	ts.Close()
	if err := dbContainer.Terminate(ctx); err != nil {
		slog.Error("failed to terminate pg container", "err", err)
	}

	os.Exit(code)
}

func TestUserJourneyFlow(t *testing.T) {
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}
	defer page.Close()

	username := fmt.Sprintf("e2etest%d", time.Now().Unix())
	email := fmt.Sprintf("e2e_%d@example.com", time.Now().Unix())
	password := "Password123!"

	t.Run("Signup", func(t *testing.T) {
		if _, err = page.Goto(ts.URL + "/signup"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		err = page.Locator("form[hx-post='/api/v1/auth/signup']").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to find signup form: %v", err)
		}

		_ = page.Locator("input[name='name']").Fill("E2E Test User")
		_ = page.Locator("input[name='username']").Fill(username)
		_ = page.Locator("input[name='email']").Fill(email)
		_ = page.Locator("input[name='password']").Fill(password)
		_ = page.Locator("input[id='confirm-password']").Fill(password)

		_ = page.Locator("button[type='submit']").Click()

		err = page.WaitForURL(ts.URL+"/dashboard", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			errText, _ := page.Locator(".text-error").TextContent()
			t.Fatalf("expected redirect to dashboard, got error: %s", errText)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		// Click on logout button
		err := page.Locator("a[hx-post='/api/v1/auth/logout']").Click()
		if err != nil {
			t.Fatalf("failed to click logout: %v", err)
		}

		err = page.WaitForURL(ts.URL+"/login", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected redirect to login: %v", err)
		}
	})

	t.Run("Login", func(t *testing.T) {
		err = page.Locator("form[hx-post='/api/v1/auth/login']").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to find login form: %v", err)
		}

		_ = page.Locator("input[name='username']").Fill(username)
		_ = page.Locator("input[name='password']").Fill(password)

		_ = page.Locator("button[type='submit']").Click()

		err = page.WaitForURL(ts.URL+"/dashboard", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			errText, _ := page.Locator(".text-error").TextContent()
			t.Fatalf("expected redirect to dashboard, got error: %s", errText)
		}
	})

	testUrl := "https://example.com/very/long/url/for/testing"

	t.Run("Create URL", func(t *testing.T) {
		_ = page.Locator("input[name='originalUrl']").Fill(testUrl)

		_ = page.Locator("form[hx-post='/dashboard/urls/shorten'] button[type='submit']").Click()

		// Wait for the URL list to update
		err = page.Locator(fmt.Sprintf("text=%s", testUrl)).WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to see newly created URL: %v", err)
		}
	})

	t.Run("Delete URL", func(t *testing.T) {
		// Wait for the delete button
		deleteBtn := page.Locator("button[title='Delete']").First()

		err = deleteBtn.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to find delete button: %v", err)
		}

		// Click the delete button to reveal confirmation
		_ = deleteBtn.Click()

		yesBtn := page.Locator("button:has-text('YES')").First()
		err = yesBtn.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(3000),
		})
		if err != nil {
			t.Fatalf("failed to find confirmation YES button: %v", err)
		}

		_ = yesBtn.Click()

		// Wait for the item to disappear
		err = page.Locator(fmt.Sprintf("text=%s", testUrl)).WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("URL did not disappear after deletion: %v", err)
		}
	})
}

func TestEmailVerificationFlow(t *testing.T) {
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}
	defer page.Close()

	username := fmt.Sprintf("veriftest%d", time.Now().Unix())
	email := fmt.Sprintf("verif_%d@example.com", time.Now().Unix())
	password := "Password123!"

	t.Run("Signup User", func(t *testing.T) {
		if _, err = page.Goto(ts.URL + "/signup"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		err = page.Locator("form[hx-post='/api/v1/auth/signup']").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to find signup form: %v", err)
		}

		_ = page.Locator("input[name='name']").Fill("Unverified User")
		_ = page.Locator("input[name='username']").Fill(username)
		_ = page.Locator("input[name='email']").Fill(email)
		_ = page.Locator("input[name='password']").Fill(password)
		_ = page.Locator("input[id='confirm-password']").Fill(password)

		_ = page.Locator("button[type='submit']").Click()

		err = page.WaitForURL(ts.URL+"/dashboard", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected redirect to dashboard: %v", err)
		}
	})

	t.Run("Logout User", func(t *testing.T) {
		err := page.Locator("a[hx-post='/api/v1/auth/logout']").Click()
		if err != nil {
			t.Fatalf("failed to click logout: %v", err)
		}

		err = page.WaitForURL(ts.URL+"/login", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected redirect to login: %v", err)
		}
	})

	t.Run("Enable Mail And Try Login", func(t *testing.T) {
		origEnv := config.GetEnvConfig()
		testEnv := *origEnv
		testEnv.MAIL_ENABLED = true
		config.SetEnvConfigForTest(&testEnv)
		defer config.SetEnvConfigForTest(origEnv)

		err = page.Locator("form[hx-post='/api/v1/auth/login']").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("failed to find login form: %v", err)
		}

		_ = page.Locator("input[name='username']").Fill(username)
		_ = page.Locator("input[name='password']").Fill(password)
		_ = page.Locator("button[type='submit']").Click()

		// Should see "Your email is not verified." message
		err = page.Locator("text=Your email is not verified.").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected 'Your email is not verified.' alert: %v", err)
		}

		// Click on "Verify your email" link
		err = page.Locator("a:has-text('Verify your email')").Click()
		if err != nil {
			t.Fatalf("failed to click verify email link: %v", err)
		}

		// Wait for verify email prompt page
		err = page.WaitForURL(fmt.Sprintf("%s/verify-email?email=%s", ts.URL, email), playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected redirect to verify-email prompt: %v", err)
		}

		// Submit resend verification email form
		_ = page.Locator("button[type='submit']").Click()

		// Wait for success alert
		err = page.Locator(".alert-success").WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("expected success alert after resend verification: %v", err)
		}
	})
}
