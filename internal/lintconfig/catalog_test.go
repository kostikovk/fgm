package lintconfig

import (
	"testing"
)

func TestLoadCatalog(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if len(catalog.Linters) == 0 {
		t.Fatal("expected linters, got none")
	}
	if len(catalog.Conflicts) == 0 {
		t.Fatal("expected conflicts, got none")
	}
}

func TestCatalog_Lookup(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	l, ok := catalog.Lookup("govet")
	if !ok {
		t.Fatal("govet not found")
	}
	if l.Category != "bugs" {
		t.Fatalf("govet category = %q, want %q", l.Category, "bugs")
	}

	_, ok = catalog.Lookup("nonexistent")
	if ok {
		t.Fatal("nonexistent should not be found")
	}
}

func TestLinter_Available(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		linter   Linter
		goMinor  int
		expected bool
	}{
		{"no constraint", Linter{MinGoMinor: 0}, 20, true},
		{"meets minimum", Linter{MinGoMinor: 22}, 22, true},
		{"exceeds minimum", Linter{MinGoMinor: 22}, 23, true},
		{"below minimum", Linter{MinGoMinor: 22}, 21, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.linter.Available(tt.goMinor); got != tt.expected {
				t.Fatalf("Available(%d) = %v, want %v", tt.goMinor, got, tt.expected)
			}
		})
	}
}

func TestCatalog_ConflictsWith(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	msg := catalog.ConflictsWith("gofumpt", "gofmt")
	if msg == "" {
		t.Fatal("expected conflict between gofumpt and gofmt")
	}

	msg = catalog.ConflictsWith("gofmt", "gofumpt")
	if msg == "" {
		t.Fatal("expected conflict in reverse order")
	}

	msg = catalog.ConflictsWith("govet", "errcheck")
	if msg != "" {
		t.Fatalf("unexpected conflict: %s", msg)
	}
}

func TestParseGoMinor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1.22.5", 22, false},
		{"1.21", 21, false},
		{"go1.23.1", 23, false},
		{"invalid", 0, true},
		{"1.abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseGoMinor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseGoMinor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseGoMinor(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
