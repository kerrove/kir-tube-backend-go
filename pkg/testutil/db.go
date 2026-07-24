// Package testutil provides shared test helpers: a real-Postgres harness (gated
// so plain `go test -short` skips it), an in-memory object store, a test config
// and row fixtures. It deliberately does NOT import internal/app so that
// white-box test packages can use it without an import cycle; the e2e harness
// builds the app via app.Assemble directly from its own test file.
package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"go/kir-tube/configs"
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/history"
	"go/kir-tube/internal/playlist"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/jwt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSecret is the JWT signing secret used across the test suite.
const TestSecret = "test-secret"

// advisoryLockKey serializes database-backed tests across processes. `go test`
// runs different packages' test binaries in parallel against the single shared
// test database, which would otherwise race on AutoMigrate and clobber each
// other's rows. A session-level Postgres advisory lock, held for the whole test
// on a dedicated connection, makes DB tests mutually exclusive globally.
const advisoryLockKey int64 = 0x6b6972747562 // "kirtub"

// defaultTestDSN points at the kir-tube_test database on the local Postgres from
// docker-compose. Override with the TEST_DSN environment variable.
const defaultTestDSN = "host=localhost user=postgres password=55Kirill55 dbname=kir-tube_test port=5432 sslmode=disable"

func testDSN() string {
	if dsn := os.Getenv("TEST_DSN"); dsn != "" {
		return dsn
	}
	return defaultTestDSN
}

// migrated is the full set of models the test database needs.
func migratedModels() []any {
	return []any{
		&user.User{},
		&channel.Channel{},
		&video.Video{},
		&video.VideoTag{},
		&video.VideoComment{},
		&video.VideoLike{},
		&playlist.Playlist{},
		&history.WatchHistory{},
	}
}

// RequireDB returns a connection to the test database, migrating the schema and
// truncating every table before and after the test. It skips (never fails) when
// running with -short or when the database is unreachable, so unit-only runs need
// no infrastructure.
func RequireDB(t *testing.T) *db.Db {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping database-backed test in -short mode")
	}

	gdb, err := gorm.Open(postgres.Open(testDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("test database unavailable (%v); start docker-compose and create kir-tube_test", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		t.Skipf("test database unavailable: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("test database unreachable (%v); start docker-compose and create kir-tube_test", err)
	}

	// Hold a session-level advisory lock on a dedicated connection for the whole
	// test so concurrent package test binaries do not migrate or mutate the
	// shared database at the same time.
	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Skipf("test database connection unavailable: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		conn.Close()
		t.Fatalf("acquire advisory lock: %v", err)
	}
	// LIFO cleanup: releasing the lock must run last, after TruncateAll below.
	t.Cleanup(func() {
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockKey)
		_ = conn.Close()
	})

	database := &db.Db{DB: gdb}
	if err := gdb.AutoMigrate(migratedModels()...); err != nil {
		t.Fatalf("automigrate test database: %v", err)
	}

	TruncateAll(t, database)
	t.Cleanup(func() { TruncateAll(t, database) })

	return database
}

// TruncateAll empties every table in the public schema, resetting identities and
// cascading through foreign keys, so each test starts from a clean slate.
func TruncateAll(t *testing.T, database *db.Db) {
	t.Helper()

	var tables []string
	if err := database.Raw(
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public'",
	).Scan(&tables).Error; err != nil {
		t.Fatalf("list tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	quoted := make([]string, len(tables))
	for i, name := range tables {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	stmt := "TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE"
	if err := database.Exec(stmt).Error; err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// TestConfig returns a config wired for tests: the test secret, non-secure
// cookies and a fixed client URL. The DB DSN points at the test database.
func TestConfig() *configs.Config {
	return &configs.Config{
		Db: configs.DbConfig{Dsn: testDSN()},
		Auth: configs.AuthConfig{
			Secret:        TestSecret,
			SecureCookies: false,
		},
		Network: configs.NetworkConfig{
			Port:      "0",
			Domain:    "localhost",
			ClientUrl: "http://localhost:3000",
		},
	}
}

// IssueToken mints a signed access token for the given user id, matching what the
// auth service produces, so tests can call authenticated endpoints.
func IssueToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := jwt.NewJWT(TestSecret).Create(jwt.JWTData{Id: userID, IsAdmin: false}, "1h")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return token
}
