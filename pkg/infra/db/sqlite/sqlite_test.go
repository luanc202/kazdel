package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"kazdel/pkg/entity"
	"kazdel/pkg/infra/db/sqlite"
	"kazdel/pkg/uniqueEntityId"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := "file:test_sqlite.db?cache=shared&mode=rwc"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}

	driver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		db.Close()
		t.Fatalf("failed to create sqlite migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://../../../../migrations/sqlite", "sqlite", driver)
	if err != nil {
		db.Close()
		t.Fatalf("failed to init migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		db.Close()
		t.Fatalf("failed to run migrations up: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	db.Close()
	os.Remove("test_sqlite.db")
}

func TestSQLiteRepositories(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// 1. Test UserRepository
	userRepo := sqlite.NewUserRepository(db)

	userId := uniqueEntityId.NewID()
	now := time.Now().UTC().Truncate(time.Second)
	user := &entity.User{
		ID:            userId,
		Name:          "Test User",
		Username:      "testuser",
		Email:         "sqlite@example.com",
		PasswordHash:  "hashed_secret",
		Role:          entity.RoleUser,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := userRepo.Save(user)
	if err != nil {
		t.Fatalf("UserRepository.Save failed: %v", err)
	}

	foundUser, err := userRepo.FindById(userId.String())
	if err != nil {
		t.Fatalf("UserRepository.FindById failed: %v", err)
	}
	if foundUser.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, foundUser.Email)
	}

	foundByEmail, err := userRepo.FindByEmail(user.Email)
	if err != nil {
		t.Fatalf("UserRepository.FindByEmail failed: %v", err)
	}
	if foundByEmail.ID != userId {
		t.Errorf("expected id %s, got %s", userId, foundByEmail.ID)
	}

	exists, err := userRepo.ExistsByEmail(user.Email)
	if err != nil || !exists {
		t.Errorf("expected exists by email true, got %v, err: %v", exists, err)
	}

	// 2. Test ShortenedUrlRepository
	urlRepo := sqlite.NewShortenedUrlRepository(db)
	shortUrl := &entity.ShortenedUrl{
		ID:        1,
		LongUrl:   "https://example.com/long-url",
		ShortSlug: "sqltst",
		UserId:    userId,
		Views:     0,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	err = urlRepo.Save(shortUrl)
	if err != nil {
		t.Fatalf("ShortenedUrlRepository.Save failed: %v", err)
	}

	foundUrl, err := urlRepo.FindBySlug("sqltst")
	if err != nil {
		t.Fatalf("ShortenedUrlRepository.FindBySlug failed: %v", err)
	}
	if foundUrl.LongUrl != shortUrl.LongUrl {
		t.Errorf("expected original url %s, got %s", shortUrl.LongUrl, foundUrl.LongUrl)
	}

	userUrls, total, err := urlRepo.FindByUserIdPaginated(userId, "", 1, 10)
	if err != nil {
		t.Fatalf("ShortenedUrlRepository.FindByUserIdPaginated failed: %v", err)
	}
	if len(userUrls) != 1 || total != 1 {
		t.Errorf("expected 1 user url, got %d (total %d)", len(userUrls), total)
	}

	// 3. Test UrlVisitRepository
	visitRepo := sqlite.NewUrlVisitRepository(db)
	ip := "127.0.0.1"
	ua := "GoTest"
	ref := "Direct"
	country := "US"
	browser := "TestBrowser"
	osName := "Linux"
	visit := &entity.UrlVisit{
		ID:        1,
		UrlId:     shortUrl.ID,
		IpAddress: &ip,
		UserAgent: &ua,
		Referrer:  &ref,
		Country:   &country,
		Browser:   &browser,
		Os:        &osName,
		ClickedAt: now,
	}
	err = visitRepo.Save(ctx, visit)
	if err != nil {
		t.Fatalf("UrlVisitRepository.Save failed: %v", err)
	}

	stats, err := visitRepo.GetStatsByUrlId(ctx, shortUrl.ID)
	if err != nil {
		t.Fatalf("UrlVisitRepository.GetStatsByUrlId failed: %v", err)
	}
	if stats.TotalClicks != 1 {
		t.Errorf("expected 1 total click, got %d", stats.TotalClicks)
	}

	// 4. Test SessionRepository
	sessionRepo := sqlite.NewSessionRepository(db)
	session := &entity.Session{
		ID:        uniqueEntityId.NewID(),
		UserID:    userId,
		Token:     "refresh-token-123",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	err = sessionRepo.Create(session)
	if err != nil {
		t.Fatalf("SessionRepository.Create failed: %v", err)
	}

	foundSession, err := sessionRepo.FindByToken("refresh-token-123")
	if err != nil {
		t.Fatalf("SessionRepository.FindByToken failed: %v", err)
	}
	if foundSession.Token != "refresh-token-123" {
		t.Errorf("expected token refresh-token-123, got %s", foundSession.Token)
	}

	// 5. Test UserTokenRepository
	tokenRepo := sqlite.NewUserTokenRepository(db)
	token := &entity.UserToken{
		ID:        uniqueEntityId.NewID(),
		UserID:    userId,
		Token:     "token_456",
		Context:   entity.TokenContextEmailVerification,
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}
	err = tokenRepo.Save(token)
	if err != nil {
		t.Fatalf("UserTokenRepository.Save failed: %v", err)
	}

	foundToken, err := tokenRepo.FindByToken("token_456")
	if err != nil {
		t.Fatalf("UserTokenRepository.FindByToken failed: %v", err)
	}
	if foundToken.Token != "token_456" {
		t.Errorf("expected token token_456, got %s", foundToken.Token)
	}

	// Clean up
	err = urlRepo.Delete(shortUrl.ID, userId)
	if err != nil {
		t.Fatalf("ShortenedUrlRepository.Delete failed: %v", err)
	}
}
