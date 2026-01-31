package greylist_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam/greylist"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

func TestGreylistFlow(t *testing.T) {
	// 1. Setup DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer conn.DB.Close()

	// 2. Setup Schema (Manually applying migration 006)
	schema := `
	CREATE TABLE greylist (
		ip_net TEXT NOT NULL,
		sender TEXT NOT NULL,
		recipient TEXT NOT NULL,
		first_seen_unix INTEGER NOT NULL,
		last_seen_unix INTEGER NOT NULL,
		blocked_count INTEGER DEFAULT 0,
		PRIMARY KEY (ip_net, sender, recipient)
	) WITHOUT ROWID;
	`
	if _, err := conn.DB.Exec(schema); err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}

	// 3. Init Service
	repo := sqlite.NewGreylistRepository(conn.DB)
	cfg := config.GreylistConfig{
		Enabled:    true,
		RetryDelay: "1s", // Short delay for testing
		Expiration: "10s",
	}

	svc, err := greylist.NewService(repo, cfg)
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	ctx := context.Background()
	tuple := domain.GreylistTuple{
		IPNet:     "1.2.3.0", // Already normalized for test
		Sender:    "spammer@example.com",
		Recipient: "victim@local",
	}

	// 4. Test Cases

	// A. First Check -> Should Fail (New)
	err = svc.Check(ctx, tuple)
	if err == nil {
		t.Fatal("expected error (greylisted) on first check, got nil")
	}
	t.Logf("First check rejected execution as expected: %v", err)

	// B. Immediate Retry -> Should Fail (Too Soon)
	err = svc.Check(ctx, tuple)
	if err == nil {
		t.Fatal("expected error (greylisted) on immediate retry, got nil")
	}
	t.Logf("Immediate retry rejected execution as expected: %v", err)

	// C. Wait for RetryDelay
	t.Log("Waiting for retry delay (1.1s)...")
	time.Sleep(1100 * time.Millisecond)

	// D. Check after Delay -> Should Pass
	err = svc.Check(ctx, tuple)
	if err != nil {
		t.Fatalf("expected pass after delay, got error: %v", err)
	}
	t.Log("Check after delay passed as expected")

	// E. Verify LastSeen updated
	// Requires direct DB check or assumption that subsequent check uses new last_seen.

	// F. Prune Check (Manual invocation)
	// Add an expired entry directly
	expiredTuple := domain.GreylistTuple{IPNet: "4.4.4.0", Sender: "old", Recipient: "old"}
	oldTime := time.Now().Add(-20 * time.Second).Unix()
	expiredEntry := &domain.GreylistEntry{
		Tuple:        expiredTuple,
		FirstSeenAt:  oldTime,
		LastSeenAt:   oldTime,
		BlockedCount: 1,
	}
	if err := repo.Upsert(ctx, expiredEntry); err != nil {
		t.Fatalf("failed to insert expired: %v", err)
	}

	count, err := svc.Prune(ctx)
	if err != nil {
		t.Fatalf("failed to prune: %v", err)
	}
	if count < 1 {
		t.Errorf("expected pruned count >= 1, got %d", count)
	}

	// Verify expired is gone
	foundEntry, err := repo.Get(ctx, expiredTuple)
	if err != nil {
		t.Fatalf("unexpected error checking pruned entry: %v", err)
	}
	if foundEntry != nil {
		t.Fatalf("expected expired entry to be gone, but found it: %+v", foundEntry)
	}
}
