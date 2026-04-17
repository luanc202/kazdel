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
	pw      *playwright.Playwright
	browser playwright.Browser
	ts      *httptest.Server
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

	// 3. Initialize app configs
	_, err = config.LoadEnv(".") // Now at project root
	if err != nil {
		fmt.Printf("Note: %v\n", err)
	}
	
	err = config.InitConfigs()
	if err != nil {
		log.Fatalf("could not initialize configs: %v", err)
	}

	// 4. Run migrations
	mi, err := migrate.New("file://migrations", connStr)
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
	sessionRepo := db.NewPostgresSessionRepository(dbConn)

	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo)

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

func TestSignupFlow(t *testing.T) {
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}
	defer page.Close()

	// 1. Navigate to signup page
	if _, err = page.Goto(ts.URL + "/signup"); err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	// Wait for the signup form to be ready
	err = page.Locator("form[hx-post='/api/v1/auth/signup']").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("failed to find signup form: %v", err)
	}

	// Fill the form (using Alpine.js client-side validation logic from previous discussions)
	_ = page.Locator("input[name='name']").Fill("E2E Test User")
	_ = page.Locator("input[name='username']").Fill(fmt.Sprintf("e2etest%d", time.Now().Unix()))
	_ = page.Locator("input[name='email']").Fill(fmt.Sprintf("e2e_%d@example.com", time.Now().Unix()))
	_ = page.Locator("input[name='password']").Fill("Password123!")
	_ = page.Locator("input[id='confirm-password']").Fill("Password123!")

	// Submit it
	_ = page.Locator("button[type='submit']").Click()

	// For this test, we expect HTMX to redirect us to the dashboard on success.
	err = page.WaitForURL(ts.URL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Logf("Failed to redirect to dashboard: %v", err)
		// Try to read error msg to debug
		errText, _ := page.Locator(".text-error").TextContent()
		t.Errorf("expected redirect, maybe got error: %s", errText)
	}
}
