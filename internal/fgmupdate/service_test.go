package fgmupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestServiceCheckLatest_FetchesCachesAndDetectsNewerVersion(t *testing.T) {
	t.Parallel()

	t.Run("should fetch latest release from GitHub API, write cache, and report update available when current version is older than remote", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tag_name":"v0.6.0"}`))
		}))
		defer srv.Close()

		fixedNow := time.Date(2026, 10, 14, 19, 30, 0, 0, time.UTC)
		svc := New(Config{
			Root:           t.TempDir(),
			CurrentVersion: "0.5.0",
			ExecutablePath: "/opt/homebrew/bin/fgm",
			Client:         srv.Client(),
			BaseURL:        srv.URL,
			Now:            func() time.Time { return fixedNow },
		})

		// Act
		status, ok, err := svc.CheckLatest(context.Background())

		// Assert
		if err != nil {
			t.Fatalf("CheckLatest: %v", err)
		}
		if !ok {
			t.Fatal("expected ok=true when version check succeeds")
		}
		if !status.UpdateAvailable {
			t.Fatal("expected UpdateAvailable=true for 0.5.0 -> 0.6.0")
		}
		if status.CurrentVersion != "0.5.0" || status.LatestVersion != "0.6.0" {
			t.Fatalf("status = %+v, want 0.5.0 -> 0.6.0", status)
		}
		if status.UpgradeCommand != "brew update && brew upgrade --cask fgm" {
			t.Fatalf("UpgradeCommand = %q, want brew upgrade command", status.UpgradeCommand)
		}
		if hits.Load() != 1 {
			t.Fatalf("hits = %d, want 1", hits.Load())
		}
	})

	t.Run("should return cached result without hitting remote endpoint when cache is within 24h TTL", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tag_name":"v0.6.0"}`))
		}))
		defer srv.Close()

		root := t.TempDir()
		fixedNow := time.Date(2026, 10, 14, 19, 30, 0, 0, time.UTC)
		svc := New(Config{
			Root:           root,
			CurrentVersion: "0.5.0",
			ExecutablePath: "/Users/dev/.local/bin/fgm",
			Client:         srv.Client(),
			BaseURL:        srv.URL,
			Now:            func() time.Time { return fixedNow },
		})

		_, _, _ = svc.CheckLatest(context.Background())

		// Act — second call 2 hours later (within 24h TTL)
		svcLater := New(Config{
			Root:           root,
			CurrentVersion: "0.5.0",
			ExecutablePath: "/Users/dev/.local/bin/fgm",
			Client:         srv.Client(),
			BaseURL:        srv.URL,
			Now:            func() time.Time { return fixedNow.Add(2 * time.Hour) },
		})
		status, ok, err := svcLater.CheckLatest(context.Background())

		// Assert
		if err != nil {
			t.Fatalf("second CheckLatest: %v", err)
		}
		if !ok || !status.UpdateAvailable {
			t.Fatalf("status = %+v, ok = %v; want cached update available", status, ok)
		}
		if hits.Load() != 1 {
			t.Fatalf("hits = %d, want 1 (cached)", hits.Load())
		}

		// Also verify CachedNotice reads disk without network
		cached, hasNotice := svcLater.CachedNotice(context.Background())
		if !hasNotice || cached.LatestVersion != "0.6.0" {
			t.Fatalf("CachedNotice = %+v, %v; want 0.6.0 notice", cached, hasNotice)
		}
	})

	t.Run("should skip check when current version is dev", func(t *testing.T) {
		t.Parallel()

		// Arrange
		svc := New(Config{
			Root:           t.TempDir(),
			CurrentVersion: "dev",
		})

		// Act
		_, ok, err := svc.CheckLatest(context.Background())

		// Assert
		if err != nil {
			t.Fatalf("CheckLatest: %v", err)
		}
		if ok {
			t.Fatal("expected ok=false when CurrentVersion is dev")
		}
	})
}
