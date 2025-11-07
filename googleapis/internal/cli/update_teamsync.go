package cli

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TeamInfoForUpdate represents a team for updating teamsync field.
type TeamInfoForUpdate struct {
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

// UpdateTeamSyncField updates the data/team.yaml file to add teamsync field.
func UpdateTeamSyncField(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Unmarshal into slice of teams
	var teams []TeamInfoForUpdate
	if err := yaml.Unmarshal(data, &teams); err != nil {
		return fmt.Errorf("unmarshaling YAML: %w", err)
	}

	// Update teams with TeamSync in description
	updated := 0
	for i := range teams {
		if strings.Contains(teams[i].Description, "TeamSync managed team") {
			teams[i].TeamSync = true
			updated++
		}
	}

	fmt.Printf("Updated %d teams with teamsync: true\n", updated)

	// Marshal back to YAML
	output, err := yaml.Marshal(teams)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Successfully updated %s\n", filePath)
	return nil
}
