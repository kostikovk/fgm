package pinnedlint

import (
	"strings"

	"github.com/koskosovu4/fgm/internal/fgmconfig"
)

// ResolvePinned returns the pinned golangci-lint version from the nearest .fgm.toml.
func ResolvePinned(workDir string) (string, bool, error) {
	config, found, err := fgmconfig.LoadNearest(workDir)
	if err != nil || !found {
		return "", false, err
	}

	version := strings.TrimSpace(config.File.Toolchain.GolangCILint)
	if version == "" || version == "auto" {
		return "", false, nil
	}

	return version, true, nil
}
