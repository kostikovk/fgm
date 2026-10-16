package versionutil

import (
	"strconv"
	"strings"
)

// CompareVersions compares two dotted version strings numerically.
// Returns negative if left < right, positive if left > right, 0 if equal.
func CompareVersions(left string, right string) int {
	leftParts := parseVersionParts(left)
	rightParts := parseVersionParts(right)

	for idx := 0; idx < len(leftParts) && idx < len(rightParts); idx++ {
		if leftParts[idx] > rightParts[idx] {
			return 1
		}
		if leftParts[idx] < rightParts[idx] {
			return -1
		}
	}

	if len(leftParts) > len(rightParts) {
		return 1
	}
	if len(leftParts) < len(rightParts) {
		return -1
	}
	return 0
}

func parseVersionParts(version string) []int {
	fields := strings.Split(version, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			return parts
		}
		parts = append(parts, value)
	}
	return parts
}

// ToolchainVersion maps a Go language version to the release that implements it.
//
// Since Go 1.21 a go directive may name a language version such as "1.26". No release is named
// go1.26; Go orders 1.26 < 1.26rc1 < 1.26.0, so the lowest stable release satisfying it is
// 1.26.0. Earlier minors ("1.20") are real release names and stay unchanged.
func ToolchainVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 2 || parts[0] != "1" {
		return version
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 21 {
		return version
	}
	return version + ".0"
}
