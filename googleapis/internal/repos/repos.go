// Package repos provides functionality to catalog repositories in a GitHub organization.
package repos

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"gopkg.in/yaml.v3"
)

type RepositoryInfo struct {
	Name           string    `yaml:"name"`
	FullName       string    `yaml:"full_name"`
	Description    string    `yaml:"description"`
	URL            string    `yaml:"url"`
	HTMLURL        string    `yaml:"html_url"`
	Language       string    `yaml:"language"`
	DefaultBranch  string    `yaml:"default_branch"`
	CreatedAt      time.Time `yaml:"created_at"`
	PushedAt       time.Time `yaml:"pushed_at"`
	UpdatedAt      time.Time `yaml:"updated_at"`
	StarCount      int       `yaml:"star_count"`
	ForkCount      int       `yaml:"fork_count"`
	OpenIssues     int       `yaml:"open_issues"`
	HasIssues      bool      `yaml:"has_issues"`
	HasProjects    bool      `yaml:"has_projects"`
	HasWiki        bool      `yaml:"has_wiki"`
	Archived       bool      `yaml:"archived"`
	Disabled       bool      `yaml:"disabled"`
	Private        bool      `yaml:"private"`
	License        string    `yaml:"license"`
	Topics         []string  `yaml:"topics"`
	Visibility     string    `yaml:"visibility"`
	Size           int       `yaml:"size"`
	HasCodeowners  bool      `yaml:"has_codeowners"`
	HasCI          bool      `yaml:"has_ci"`
	Classification string    `yaml:"classification"`
	DaysSincePush  int       `yaml:"days_since_push"`
}

func getGitHubToken() (string, error) {
	// Try to get token from gh CLI
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from gh CLI: %w (make sure you're authenticated with 'gh auth login')", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func RunSingle(org, repoName, output string) error {
	token, err := getGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(token)

	log.Printf("Cataloging repository %s in %s organization...\n", repoName, org)

	// Fetch the specific repository
	repo, _, err := client.Repositories.Get(ctx, org, repoName)
	if err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	info := RepositoryInfo{
		Name:          repo.GetName(),
		FullName:      repo.GetFullName(),
		Description:   repo.GetDescription(),
		URL:           repo.GetURL(),
		HTMLURL:       repo.GetHTMLURL(),
		Language:      repo.GetLanguage(),
		DefaultBranch: repo.GetDefaultBranch(),
		CreatedAt:     repo.GetCreatedAt().Time,
		PushedAt:      repo.GetPushedAt().Time,
		UpdatedAt:     repo.GetUpdatedAt().Time,
		StarCount:     repo.GetStargazersCount(),
		ForkCount:     repo.GetForksCount(),
		OpenIssues:    repo.GetOpenIssuesCount(),
		HasIssues:     repo.GetHasIssues(),
		HasProjects:   repo.GetHasProjects(),
		HasWiki:       repo.GetHasWiki(),
		Archived:      repo.GetArchived(),
		Disabled:      repo.GetDisabled(),
		Private:       repo.GetPrivate(),
		Topics:        repo.Topics,
		Visibility:    repo.GetVisibility(),
		Size:          repo.GetSize(),
	}

	if license := repo.GetLicense(); license != nil {
		info.License = license.GetSPDXID()
	}

	// Calculate days since last push
	if !info.PushedAt.IsZero() {
		info.DaysSincePush = int(time.Since(info.PushedAt).Hours() / 24)
	}

	// Classify repository based on activity
	info.Classification = classifyRepository(info)

	// Enrich repository data
	log.Println("Enriching repository data...")
	enrichRepository(ctx, client, org, &info)

	// Save to file
	if err := saveRepositories([]RepositoryInfo{info}, output); err != nil {
		return fmt.Errorf("failed to save repository: %w", err)
	}

	log.Printf("Repository catalog saved to %s\n", output)
	printSummary([]RepositoryInfo{info})
	return nil
}

func RunAll(org, output string) error {
	token, err := getGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(token)

	log.Printf("Cataloging repositories in %s organization...\n", org)

	repos, err := fetchRepositories(ctx, client, org)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	log.Printf("Found %d repositories\n", len(repos))

	// Enrich repository data
	log.Println("Enriching repository data...")
	for i := range repos {
		enrichRepository(ctx, client, org, &repos[i])

		// Save after each repository is processed
		if err := saveRepositories(repos, output); err != nil {
			log.Printf("Warning: failed to save repositories after processing repo %d: %v", i+1, err)
		}

		if (i+1)%10 == 0 {
			log.Printf("Processed %d/%d repositories (saved to %s)", i+1, len(repos), output)
		}
	}

	log.Printf("Repository catalog saved to %s\n", output)
	printSummary(repos)
	return nil
}

func fetchRepositories(ctx context.Context, client *github.Client, org string) ([]RepositoryInfo, error) {
	var allRepos []RepositoryInfo
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}

		for _, repo := range repos {
			info := RepositoryInfo{
				Name:          repo.GetName(),
				FullName:      repo.GetFullName(),
				Description:   repo.GetDescription(),
				URL:           repo.GetURL(),
				HTMLURL:       repo.GetHTMLURL(),
				Language:      repo.GetLanguage(),
				DefaultBranch: repo.GetDefaultBranch(),
				CreatedAt:     repo.GetCreatedAt().Time,
				PushedAt:      repo.GetPushedAt().Time,
				UpdatedAt:     repo.GetUpdatedAt().Time,
				StarCount:     repo.GetStargazersCount(),
				ForkCount:     repo.GetForksCount(),
				OpenIssues:    repo.GetOpenIssuesCount(),
				HasIssues:     repo.GetHasIssues(),
				HasProjects:   repo.GetHasProjects(),
				HasWiki:       repo.GetHasWiki(),
				Archived:      repo.GetArchived(),
				Disabled:      repo.GetDisabled(),
				Private:       repo.GetPrivate(),
				Topics:        repo.Topics,
				Visibility:    repo.GetVisibility(),
				Size:          repo.GetSize(),
			}

			if license := repo.GetLicense(); license != nil {
				info.License = license.GetSPDXID()
			}

			// Calculate days since last push
			if !info.PushedAt.IsZero() {
				info.DaysSincePush = int(time.Since(info.PushedAt).Hours() / 24)
			}

			// Classify repository based on activity
			info.Classification = classifyRepository(info)

			allRepos = append(allRepos, info)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func enrichRepository(ctx context.Context, client *github.Client, org string, repo *RepositoryInfo) {
	// Check for CODEOWNERS file
	repo.HasCodeowners = hasFile(ctx, client, org, repo.Name, "CODEOWNERS")

	// Check for CI configuration (GitHub Actions, CircleCI, Travis, etc.)
	repo.HasCI = hasFile(ctx, client, org, repo.Name, ".github/workflows") ||
		hasFile(ctx, client, org, repo.Name, ".circleci/config.yml") ||
		hasFile(ctx, client, org, repo.Name, ".travis.yml")
}

func hasFile(ctx context.Context, client *github.Client, org, repo, path string) bool {
	_, _, _, err := client.Repositories.GetContents(ctx, org, repo, path, nil)
	return err == nil
}

func classifyRepository(repo RepositoryInfo) string {
	if repo.Archived {
		return "archived"
	}

	days := repo.DaysSincePush

	switch {
	case days < 180: // 6 months
		return "active"
	case days < 730: // 24 months
		return "maintenance"
	default:
		return "stale"
	}
}

func saveRepositories(repos []RepositoryInfo, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := yaml.Marshal(repos)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func printSummary(repos []RepositoryInfo) {
	counts := map[string]int{
		"active":      0,
		"maintenance": 0,
		"stale":       0,
		"archived":    0,
	}

	languages := make(map[string]int)
	withCodeowners := 0
	withCI := 0

	for _, repo := range repos {
		counts[repo.Classification]++

		if repo.Language != "" {
			languages[repo.Language]++
		}

		if repo.HasCodeowners {
			withCodeowners++
		}

		if repo.HasCI {
			withCI++
		}
	}

	fmt.Println("\n=== Repository Summary ===")
	fmt.Printf("Total repositories: %d\n\n", len(repos))

	fmt.Println("By classification:")
	fmt.Printf("  Active (< 6 months):         %d\n", counts["active"])
	fmt.Printf("  Maintenance (6-24 months):   %d\n", counts["maintenance"])
	fmt.Printf("  Stale (> 24 months):         %d\n", counts["stale"])
	fmt.Printf("  Archived:                    %d\n\n", counts["archived"])

	fmt.Printf("With CODEOWNERS file:          %d (%.1f%%)\n", withCodeowners, float64(withCodeowners)/float64(len(repos))*100)
	fmt.Printf("With CI configuration:         %d (%.1f%%)\n", withCI, float64(withCI)/float64(len(repos))*100)

	fmt.Println("\nTop 10 languages:")
	type langCount struct {
		lang  string
		count int
	}
	var langCounts []langCount
	for lang, count := range languages {
		langCounts = append(langCounts, langCount{lang, count})
	}
	// Simple sort by count
	for i := 0; i < len(langCounts); i++ {
		for j := i + 1; j < len(langCounts); j++ {
			if langCounts[j].count > langCounts[i].count {
				langCounts[i], langCounts[j] = langCounts[j], langCounts[i]
			}
		}
	}
	for i := 0; i < 10 && i < len(langCounts); i++ {
		fmt.Printf("  %-20s %d\n", langCounts[i].lang, langCounts[i].count)
	}
}
