package resolve

import (
	"context"
	"path/filepath"
	"testing"
)

func TestResolverLanguageVersionDirective(t *testing.T) {
	t.Parallel()

	catalog := stubPatchCatalog{"1.26": "1.26.9"}

	cases := []struct {
		name     string
		files    map[string]string
		dir      string
		wantGo   string
		wantBase string
	}{
		{name: "should select 1.26.0 when go.mod declares go 1.26", files: map[string]string{"go.mod": "module m\n\ngo 1.26\n"}, wantGo: "1.26.0", wantBase: "1.26"},
		{name: "should select 1.21.0 when go.mod declares go 1.21", files: map[string]string{"go.mod": "module m\n\ngo 1.21\n"}, wantGo: "1.21.0", wantBase: "1.21"},
		{name: "should keep 1.20 when go.mod declares a pre-1.21 go 1.20", files: map[string]string{"go.mod": "module m\n\ngo 1.20\n"}, wantGo: "1.20"},
		{name: "should keep the patch when go.mod declares go 1.26.3", files: map[string]string{"go.mod": "module m\n\ngo 1.26.3\n"}, wantGo: "1.26.3"},
		{name: "should prefer the toolchain directive when go 1.26 has toolchain go1.26.4", files: map[string]string{"go.mod": "module m\n\ngo 1.26\n\ntoolchain go1.26.4\n"}, wantGo: "1.26.4"},
		{name: "should select 1.26.0 when go.work declares go 1.26", files: map[string]string{"go.work": "go 1.26\n\nuse ./m\n", "m/go.mod": "module m\n\ngo 1.26\n"}, dir: "m", wantGo: "1.26.0", wantBase: "1.26"},
		{name: "should select the newest patch when latest-patch applies to go 1.26", files: map[string]string{"go.mod": "module m\n\ngo 1.26\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"}, wantGo: "1.26.9", wantBase: "1.26"},
		{name: "should select 1.26.0 when .fgm.toml pins go 1.26", files: map[string]string{"go.mod": "module m\n\ngo 1.26\n", ".fgm.toml": "[toolchain]\ngo = \"1.26\"\n"}, wantGo: "1.26.0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := t.TempDir()
			writeFiles(t, root, tc.files)

			// Act
			selection, err := NewWithConfig(Config{PatchCatalog: catalog}).Current(context.Background(), filepath.Join(root, tc.dir))

			// Assert
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.GoVersion != tc.wantGo || selection.GoBase != tc.wantBase {
				t.Errorf("selection = %s (base %q); want %s (base %q)", selection.GoVersion, selection.GoBase, tc.wantGo, tc.wantBase)
			}
		})
	}
}

func TestValidateGoPinLanguageVersion(t *testing.T) {
	t.Parallel()

	t.Run("ValidateGoPin", func(t *testing.T) {
		t.Parallel()

		t.Run("should accept 1.26.0 when the go directive is the language version 1.26", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := t.TempDir()
			writeFiles(t, root, map[string]string{"go.mod": "module m\n\ngo 1.26\n"})

			// Act
			err := ValidateGoPin(root, "1.26.0")

			// Assert
			if err != nil {
				t.Errorf("ValidateGoPin: %v", err)
			}
		})
	})
}
