// Package teams provides functionality to catalog teams in a GitHub organization.
package teams

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v66/github"
	"gopkg.in/yaml.v3"
)

type TeamInfo struct {
	ID             int64    `yaml:"id"`
	Name           string   `yaml:"name"`
	Slug           string   `yaml:"slug"`
	Description    string   `yaml:"description"`
	Privacy        string   `yaml:"privacy"`
	URL            string   `yaml:"url"`
	HTMLURL        string   `yaml:"html_url"`
	MemberCount    int      `yaml:"member_count"`
	RepoCount      int      `yaml:"repo_count"`
	ParentTeamName string   `yaml:"parent_team_name,omitempty"`
	ParentTeamID   int64    `yaml:"parent_team_id,omitempty"`
	Members        []string `yaml:"members"`
	Repositories   []string `yaml:"repositories"`
	Classification string   `yaml:"classification"`
	TeamSync       bool     `yaml:"teamsync"`
}

type TeamMemberInfo struct {
	Login string `yaml:"login"`
	Role  string `yaml:"role"`
}

type TeamRepoInfo struct {
	Name       string `yaml:"name"`
	Permission string `yaml:"permission"`
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

func RunSingle(org, teamSlug, output string) error {
	token, err := getGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(token)

	log.Printf("Cataloging team %s in %s organization...\n", teamSlug, org)

	// Fetch the specific team
	team, _, err := client.Teams.GetTeamBySlug(ctx, org, teamSlug)
	if err != nil {
		return fmt.Errorf("failed to fetch team: %w", err)
	}

	info := TeamInfo{
		ID:          team.GetID(),
		Name:        team.GetName(),
		Slug:        team.GetSlug(),
		Description: team.GetDescription(),
		Privacy:     team.GetPrivacy(),
		URL:         team.GetURL(),
		HTMLURL:     team.GetHTMLURL(),
	}

	if parent := team.Parent; parent != nil {
		info.ParentTeamID = parent.GetID()
		info.ParentTeamName = parent.GetName()
	}

	// Enrich team data
	log.Println("Enriching team data...")
	enrichTeam(ctx, client, org, &info)

	// Save to file
	if err := saveTeams([]TeamInfo{info}, output); err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	log.Printf("Team catalog saved to %s\n", output)
	printSummary([]TeamInfo{info})
	return nil
}

func RunAll(org, output string) error {
	token, err := getGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(token)

	log.Printf("Cataloging teams in %s organization...\n", org)

	teams, err := fetchTeams(ctx, client, org)
	if err != nil {
		return fmt.Errorf("failed to fetch teams: %w", err)
	}

	log.Printf("Found %d teams\n", len(teams))

	// Enrich team data
	log.Println("Enriching team data...")
	for i := range teams {
		enrichTeam(ctx, client, org, &teams[i])

		// Save after each team is processed
		if err := saveTeams(teams, output); err != nil {
			log.Printf("Warning: failed to save teams after processing team %d: %v", i+1, err)
		}

		if (i+1)%10 == 0 {
			log.Printf("Processed %d/%d teams (saved to %s)", i+1, len(teams), output)
		}
	}

	log.Printf("Team catalog saved to %s\n", output)
	printSummary(teams)
	return nil
}

func fetchTeams(ctx context.Context, client *github.Client, org string) ([]TeamInfo, error) {
	var allTeams []TeamInfo
	opts := &github.ListOptions{PerPage: 100}

	for {
		teams, resp, err := client.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("listing teams: %w", err)
		}

		for _, team := range teams {
			info := TeamInfo{
				ID:          team.GetID(),
				Name:        team.GetName(),
				Slug:        team.GetSlug(),
				Description: team.GetDescription(),
				Privacy:     team.GetPrivacy(),
				URL:         team.GetURL(),
				HTMLURL:     team.GetHTMLURL(),
			}

			if parent := team.Parent; parent != nil {
				info.ParentTeamID = parent.GetID()
				info.ParentTeamName = parent.GetName()
			}

			allTeams = append(allTeams, info)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTeams, nil
}

func isTeamSyncManaged(description string) bool {
	return strings.Contains(description, "TeamSync managed team")
}

func enrichTeam(ctx context.Context, client *github.Client, org string, team *TeamInfo) {
	// Fetch team members
	members, err := fetchTeamMembers(ctx, client, org, team.Slug)
	if err != nil {
		log.Printf("Warning: failed to fetch members for team %s: %v", team.Slug, err)
	} else {
		team.Members = members
		team.MemberCount = len(members)
	}

	// Fetch team repositories
	repos, err := fetchTeamRepositories(ctx, client, org, team.Slug)
	if err != nil {
		log.Printf("Warning: failed to fetch repositories for team %s: %v", team.Slug, err)
	} else {
		team.Repositories = repos
		team.RepoCount = len(repos)
	}

	// Check if team is managed by TeamSync
	team.TeamSync = isTeamSyncManaged(team.Description)

	// Classify team
	team.Classification = classifyTeam(*team)
}

func fetchTeamMembers(ctx context.Context, client *github.Client, org, teamSlug string) ([]string, error) {
	var members []string
	opts := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		users, resp, err := client.Teams.ListTeamMembersBySlug(ctx, org, teamSlug, opts)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			members = append(members, user.GetLogin())
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return members, nil
}

func fetchTeamRepositories(ctx context.Context, client *github.Client, org, teamSlug string) ([]string, error) {
	var repos []string
	opts := &github.ListOptions{PerPage: 100}

	for {
		repositories, resp, err := client.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
		if err != nil {
			return nil, err
		}

		for _, repo := range repositories {
			repos = append(repos, repo.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

func classifyTeam(team TeamInfo) string {
	if team.MemberCount == 0 && team.RepoCount == 0 {
		return "empty"
	}
	if team.MemberCount == 0 {
		return "no_members"
	}
	if team.RepoCount == 0 {
		return "no_repos"
	}
	return "active"
}

func saveTeams(teams []TeamInfo, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := yaml.Marshal(teams)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func printSummary(teams []TeamInfo) {
	counts := map[string]int{
		"active":     0,
		"empty":      0,
		"no_members": 0,
		"no_repos":   0,
	}

	totalMembers := 0
	totalRepos := 0
	withParent := 0

	for _, team := range teams {
		counts[team.Classification]++
		totalMembers += team.MemberCount
		totalRepos += team.RepoCount

		if team.ParentTeamID != 0 {
			withParent++
		}
	}

	fmt.Println("\n=== Team Summary ===")
	fmt.Printf("Total teams: %d\n\n", len(teams))

	fmt.Println("By classification:")
	fmt.Printf("  Active (has members and repos): %d\n", counts["active"])
	fmt.Printf("  No members:                     %d\n", counts["no_members"])
	fmt.Printf("  No repositories:                %d\n", counts["no_repos"])
	fmt.Printf("  Empty (no members or repos):    %d\n\n", counts["empty"])

	fmt.Printf("Teams with parent:                %d (%.1f%%)\n", withParent, float64(withParent)/float64(len(teams))*100)
	fmt.Printf("Total members across all teams:   %d\n", totalMembers)
	fmt.Printf("Total repos across all teams:     %d\n", totalRepos)

	if len(teams) > 0 {
		fmt.Printf("Average members per team:         %.1f\n", float64(totalMembers)/float64(len(teams)))
		fmt.Printf("Average repos per team:           %.1f\n", float64(totalRepos)/float64(len(teams)))
	}

	// Find teams that should be investigated
	fmt.Println("\n=== Cleanup Candidates ===")
	emptyCount := counts["empty"] + counts["no_members"]
	fmt.Printf("Teams with no members:            %d\n", emptyCount)
	fmt.Printf("Teams with no repos:              %d\n", counts["no_repos"])
	fmt.Printf("Potential cleanup candidates:     %d (%.1f%%)\n",
		emptyCount+counts["no_repos"]-counts["empty"],
		float64(emptyCount+counts["no_repos"]-counts["empty"])/float64(len(teams))*100)
}
