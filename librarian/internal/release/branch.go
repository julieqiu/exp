// Package release provides utilities for managing releases and branches.
package release

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/julieqiu/exp/librarian/internal/config"
)

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentCommit returns the current git commit SHA.
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// DetectPrerelease detects the prerelease suffix based on the current branch
// and configured branch patterns.
func DetectPrerelease(cfg *config.Config) (string, error) {
	if cfg.Release == nil || len(cfg.Release.BranchPatterns) == 0 {
		return "", nil
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		return "", err
	}

	// Match against patterns
	for _, pattern := range cfg.Release.BranchPatterns {
		matched, err := filepath.Match(pattern.Pattern, branch)
		if err != nil {
			continue
		}
		if matched {
			return pattern.Prerelease, nil
		}
	}

	return "", nil
}
