package db

import (
	"database/sql"
	"fmt"
	"kazdel/pkg/infra/db/postgres"
	"kazdel/pkg/infra/db/sqlite"
	interfaces "kazdel/pkg/interface"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewUserRepository(dbConn any) interfaces.UserRepository {
	switch conn := dbConn.(type) {
	case *sql.DB:
		return sqlite.NewUserRepository(conn)
	case *pgxpool.Pool:
		return postgres.NewUserRepository(conn)
	default:
		panic(fmt.Sprintf("unsupported database connection type for UserRepository: %T", dbConn))
	}
}

func NewShortenedUrlRepository(dbConn any) interfaces.ShortenedUrlRepository {
	switch conn := dbConn.(type) {
	case *sql.DB:
		return sqlite.NewShortenedUrlRepository(conn)
	case *pgxpool.Pool:
		return postgres.NewShortenedUrlRepository(conn)
	default:
		panic(fmt.Sprintf("unsupported database connection type for ShortenedUrlRepository: %T", dbConn))
	}
}

func NewUrlVisitRepository(dbConn any) interfaces.UrlVisitRepository {
	switch conn := dbConn.(type) {
	case *sql.DB:
		return sqlite.NewUrlVisitRepository(conn)
	case *pgxpool.Pool:
		return postgres.NewUrlVisitRepository(conn)
	default:
		panic(fmt.Sprintf("unsupported database connection type for UrlVisitRepository: %T", dbConn))
	}
}

func NewSessionRepository(dbConn any) interfaces.SessionRepository {
	switch conn := dbConn.(type) {
	case *sql.DB:
		return sqlite.NewSessionRepository(conn)
	case *pgxpool.Pool:
		return postgres.NewPostgresSessionRepository(conn)
	default:
		panic(fmt.Sprintf("unsupported database connection type for SessionRepository: %T", dbConn))
	}
}

func NewUserTokenRepository(dbConn any) interfaces.UserTokenRepository {
	switch conn := dbConn.(type) {
	case *sql.DB:
		return sqlite.NewUserTokenRepository(conn)
	case *pgxpool.Pool:
		return postgres.NewUserTokenRepository(conn)
	default:
		panic(fmt.Sprintf("unsupported database connection type for UserTokenRepository: %T", dbConn))
	}
}
