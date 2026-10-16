package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestAgentSkillFiles(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("skills/fgm/SKILL.md")
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}

	t.Run("skills/fgm", func(t *testing.T) {
		t.Parallel()

		t.Run("should keep the SKILL.md body under 2 KB when the skill loads", func(t *testing.T) {
			t.Parallel()

			// Arrange
			parts := bytes.SplitN(content, []byte("\n---\n"), 2)

			// Act
			body := parts[len(parts)-1]

			// Assert
			if len(parts) != 2 || len(body) > 2048 {
				t.Errorf("SKILL.md body is %d bytes (frontmatter found: %v); want at most 2048", len(body), len(parts) == 2)
			}
		})

		t.Run("should link only reference files that exist when SKILL.md points to details", func(t *testing.T) {
			t.Parallel()

			// Arrange
			links := regexp.MustCompile(`\]\((references/[^)#]+)`).FindAllSubmatch(content, -1)

			// Act
			var missing []string
			for _, link := range links {
				if _, err := os.Stat(filepath.Join("skills/fgm", string(link[1]))); err != nil {
					missing = append(missing, string(link[1]))
				}
			}

			// Assert
			if len(links) == 0 || len(missing) > 0 {
				t.Errorf("reference links = %d, missing = %v; want existing references", len(links), missing)
			}
		})

		t.Run("should mirror every skill file into .agents/skills/fgm when the skill changes", func(t *testing.T) {
			t.Parallel()

			// Arrange
			var mismatched []string

			// Act
			err := filepath.WalkDir("skills/fgm", func(path string, entry os.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				rel, _ := filepath.Rel("skills/fgm", path)
				source, _ := os.ReadFile(path)
				mirror, mirrorErr := os.ReadFile(filepath.Join(".agents/skills/fgm", rel))
				if mirrorErr != nil || !bytes.Equal(source, mirror) {
					mismatched = append(mismatched, rel)
				}
				return nil
			})

			// Assert
			if err != nil || len(mismatched) > 0 {
				t.Errorf("walk err %v; files differing from the mirror: %v", err, mismatched)
			}
		})
	})
}
