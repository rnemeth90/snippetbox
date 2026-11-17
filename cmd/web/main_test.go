package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rnemeth90/snippetbox/internal/models"
)

func TestSQLiteDatabase(t *testing.T) {
	db, err := openDB("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := initializeSchema(db, "sqlite"); err != nil {
		t.Fatal(err)
	}

	checkDatabaseModels(t, db, "sqlite")
}

func TestPostgreSQLDatabase(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set")
	}

	db, err := openDB("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := initializeSchema(db, "postgres"); err != nil {
		t.Fatal(err)
	}
	checkDatabaseModels(t, db, "postgres")
}

func checkDatabaseModels(t *testing.T, db *sql.DB, dialect string) {
	t.Helper()

	snippets := &models.SnippetModel{DB: db, Dialect: dialect}
	id, err := snippets.Insert("title", "content", 7)
	if err != nil {
		t.Fatal(err)
	}
	snippet, err := snippets.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if snippet.Title != "title" || snippet.Content != "content" {
		t.Fatalf("unexpected snippet: %#v", snippet)
	}

	users := &models.UserModel{DB: db, Dialect: dialect}
	email := fmt.Sprintf("alice-%d@example.com", time.Now().UnixNano())
	if err := users.Insert("alice", email, "valid-pass"); err != nil {
		t.Fatal(err)
	}
	if err := users.Insert("alice", email, "valid-pass"); !errors.Is(err, models.ErrDuplicateEmail) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
	if _, err := users.Authenticate(email, "valid-pass"); err != nil {
		t.Fatal(err)
	}
}
