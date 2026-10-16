package goenv

import (
	"slices"
	"strings"
	"testing"
)

func TestParseMode(t *testing.T) {
	t.Run("ParseMode", func(t *testing.T) {
		cases := []struct {
			name    string
			value   string
			want    Mode
			wantErr bool
		}{
			{name: "should select local when value is empty", value: "", want: ModeLocal},
			{name: "should select local when value is local", value: "local", want: ModeLocal},
			{name: "should select inherit when value is inherit with surrounding spaces", value: " inherit ", want: ModeInherit},
			{name: "should reject a value outside local and inherit", value: "go1.26.8", wantErr: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				value := tc.value

				// Act
				got, err := ParseMode(value)

				// Assert
				if tc.wantErr {
					if err == nil {
						t.Fatalf("ParseMode(%q) returned no error; want an error", value)
					}
					return
				}
				if err != nil {
					t.Fatalf("ParseMode(%q): %v", value, err)
				}
				if got != tc.want {
					t.Errorf("ParseMode(%q) = %q; want %q", value, got, tc.want)
				}
			})
		}
	})
}

func TestApply(t *testing.T) {
	t.Run("Apply", func(t *testing.T) {
		t.Run("should force GOTOOLCHAIN=local when an inherited version is set in local mode", func(t *testing.T) {
			// Arrange
			env := []string{"HOME=/home/u", "GOTOOLCHAIN=go1.26.8", "PATH=/bin"}

			// Act
			got, warning := Apply(env, ModeLocal)

			// Assert
			if !slices.Contains(got, "GOTOOLCHAIN=local") {
				t.Fatalf("env = %v; want GOTOOLCHAIN=local", got)
			}
			if slices.Contains(got, "GOTOOLCHAIN=go1.26.8") {
				t.Fatalf("env = %v; want inherited GOTOOLCHAIN removed", got)
			}
			if warning != "" {
				t.Errorf("warning = %q; want none", warning)
			}
		})

		t.Run("should remove GOROOT when it is inherited in local mode", func(t *testing.T) {
			// Arrange
			env := []string{"GOROOT=/opt/go1.26.8", "PATH=/bin"}

			// Act
			got, _ := Apply(env, ModeLocal)

			// Assert
			for _, entry := range got {
				if strings.HasPrefix(entry, "GOROOT=") {
					t.Fatalf("env = %v; want GOROOT removed", got)
				}
			}
		})

		t.Run("should keep inherited GOTOOLCHAIN and GOROOT when mode is inherit", func(t *testing.T) {
			// Arrange
			env := []string{"GOTOOLCHAIN=go1.26.8", "GOROOT=/opt/go1.26.8"}

			// Act
			got, _ := Apply(env, ModeInherit)

			// Assert
			if !slices.Equal(got, env) {
				t.Errorf("env = %v; want %v", got, env)
			}
		})

		t.Run("should warn when an inherited GOTOOLCHAIN names a toolchain in inherit mode", func(t *testing.T) {
			// Arrange
			env := []string{"GOTOOLCHAIN=go1.26.8"}

			// Act
			_, warning := Apply(env, ModeInherit)

			// Assert
			if !strings.Contains(warning, "GOTOOLCHAIN=go1.26.8") {
				t.Errorf("warning = %q; want it to name the inherited value", warning)
			}
		})

		t.Run("should not warn when the inherited GOTOOLCHAIN is auto in inherit mode", func(t *testing.T) {
			// Arrange
			env := []string{"GOTOOLCHAIN=auto"}

			// Act
			_, warning := Apply(env, ModeInherit)

			// Assert
			if warning != "" {
				t.Errorf("warning = %q; want none", warning)
			}
		})
	})
}

func TestOverridesSelection(t *testing.T) {
	t.Run("OverridesSelection", func(t *testing.T) {
		cases := []struct {
			name  string
			value string
			want  bool
		}{
			{name: "should report false when GOTOOLCHAIN is empty", value: "", want: false},
			{name: "should report false when GOTOOLCHAIN is local", value: "local", want: false},
			{name: "should report false when GOTOOLCHAIN is auto", value: "auto", want: false},
			{name: "should report false when GOTOOLCHAIN is path", value: "path", want: false},
			{name: "should report true when GOTOOLCHAIN names a version", value: "go1.26.8", want: true},
			{name: "should report true when GOTOOLCHAIN names a version with auto fallback", value: "go1.26.8+auto", want: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				value := tc.value

				// Act
				got := OverridesSelection(value)

				// Assert
				if got != tc.want {
					t.Errorf("OverridesSelection(%q) = %v; want %v", value, got, tc.want)
				}
			})
		}
	})
}
