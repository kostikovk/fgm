package doctor

import (
	"context"
	"strings"
	"testing"
)

type stubGoStore struct {
	globalGoVersionFn func(ctx context.Context) (string, bool, error)
	shimDirFn         func() string
}

func (s stubGoStore) GlobalGoVersion(ctx context.Context) (string, bool, error) {
	return s.globalGoVersionFn(ctx)
}

func (s stubGoStore) ShimDir() string {
	return s.shimDirFn()
}

func TestDiagnose_ReportsGlobalVersionAndShimStatus(t *testing.T) {
	t.Parallel()

	service := New(stubGoStore{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "1.25.7", true, nil
		},
		shimDirFn: func() string {
			return "/tmp/fgm/shims"
		},
	}, "/tmp/fgm/shims:/usr/local/bin")

	lines, err := service.Diagnose(context.Background())
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "OK global Go version: 1.25.7") {
		t.Fatalf("joined = %q, want global version line", joined)
	}
	if !strings.Contains(joined, "OK shim dir is on PATH: /tmp/fgm/shims") {
		t.Fatalf("joined = %q, want shim path line", joined)
	}
}

func TestDiagnose_SuggestsEnvCommandWhenShimDirMissing(t *testing.T) {
	t.Parallel()

	service := New(stubGoStore{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "", false, nil
		},
		shimDirFn: func() string {
			return "/tmp/fgm/shims"
		},
	}, "/usr/local/bin")

	lines, err := service.Diagnose(context.Background())
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `Run: eval "$(fgm env)"`) {
		t.Fatalf("joined = %q, want env suggestion", joined)
	}
}
