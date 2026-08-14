package db

import (
	"os"
	"testing"
)

func TestOpenPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		t.Skip(
			"DATABASE_URL is not configured",
		)
	}

	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatalf(
			"failed to connect to postgres: %v",
			err,
		)
	}

	defer db.Close()
}
