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
