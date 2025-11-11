package release

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-z]+)\.(\d+))?$`)

// IncrementVersion increments a version string with optional prerelease suffix.
// Examples:
//   - IncrementVersion("v1.0.0", "") -> "v1.1.0"
//   - IncrementVersion("v1.0.0", "rc") -> "v1.1.0-rc.1"
//   - IncrementVersion("v1.0.0-rc.1", "rc") -> "v1.0.0-rc.2"
func IncrementVersion(current, prerelease string) (string, error) {
	if current == "" || current == "null" {
		if prerelease != "" {
			return "v0.1.0-" + prerelease + ".1", nil
		}
		return "v0.1.0", nil
	}

	matches := versionRegex.FindStringSubmatch(current)
	if matches == nil {
		return "", fmt.Errorf("invalid version format: %s", current)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	currentPre := matches[4]
	currentPreNum := 0
	if matches[5] != "" {
		currentPreNum, _ = strconv.Atoi(matches[5])
	}

	// If adding/keeping prerelease
	if prerelease != "" {
		if currentPre == prerelease {
			// Increment prerelease number
			return fmt.Sprintf("v%d.%d.%d-%s.%d", major, minor, patch, prerelease, currentPreNum+1), nil
		}
		// New prerelease type, increment minor
		return fmt.Sprintf("v%d.%d.%d-%s.1", major, minor+1, 0, prerelease), nil
	}

	// Removing prerelease (promoting to stable)
	if currentPre != "" {
		// Just remove the prerelease suffix
		return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
	}

	// Regular increment (minor version)
	return fmt.Sprintf("v%d.%d.%d", major, minor+1, 0), nil
}

// RemovePrerelease removes the prerelease suffix from a version.
// Examples:
//   - RemovePrerelease("v1.0.0-rc.1") -> "v1.0.0"
//   - RemovePrerelease("v1.0.0") -> "v1.0.0"
func RemovePrerelease(version string) string {
	if idx := strings.Index(version, "-"); idx != -1 {
		return version[:idx]
	}
	return version
}

// HasPrerelease returns true if the version has a prerelease suffix.
func HasPrerelease(version string) bool {
	return strings.Contains(version, "-")
}
