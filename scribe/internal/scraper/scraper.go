package scraper

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Package struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}

type Library struct {
	Name     string    `yaml:"name"`
	Packages []Package `yaml:"packages"`
}

// Scrape fetches and parses Google Cloud documentation for the specified language.
func Scrape(language string) ([]Library, error) {
	url := fmt.Sprintf("https://docs.cloud.google.com/%s/docs/reference", language)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Map to group packages by service name
	libraryMap := make(map[string]*Library)

	// Libraries are organized in a table with header "Libraries"
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Check if this table has the Libraries header
		hasLibrariesHeader := false
		table.Find("th").Each(func(j int, th *goquery.Selection) {
			if strings.Contains(th.Text(), "Libraries") {
				hasLibrariesHeader = true
			}
		})

		if !hasLibrariesHeader {
			return
		}

		// Extract rows from the table body
		table.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 2 {
				return
			}

			// First cell contains the service name
			serviceName := strings.TrimSpace(cells.Eq(0).Text())
			if serviceName == "" {
				return
			}

			// Second cell contains the library links
			secondCell := cells.Eq(1)
			secondCell.Find("a").Each(func(k int, link *goquery.Selection) {
				packageName := strings.TrimSpace(link.Text())
				href, exists := link.Attr("href")
				if !exists || packageName == "" {
					return
				}

				// Make sure the link is absolute
				if strings.HasPrefix(href, "/") {
					href = "https://docs.cloud.google.com" + href
				} else if !strings.HasPrefix(href, "http") {
					// Skip relative links that aren't properly formatted
					return
				}

				// Group packages by service name
				if lib, exists := libraryMap[serviceName]; exists {
					lib.Packages = append(lib.Packages, Package{
						Name: packageName,
						Link: href,
					})
				} else {
					libraryMap[serviceName] = &Library{
						Name: serviceName,
						Packages: []Package{
							{
								Name: packageName,
								Link: href,
							},
						},
					}
				}
			})
		})
	})

	// Convert map to slice
	var libraries []Library
	for _, lib := range libraryMap {
		libraries = append(libraries, *lib)
	}

	// Sort libraries alphabetically by name
	sort.Slice(libraries, func(i, j int) bool {
		return libraries[i].Name < libraries[j].Name
	})

	return libraries, nil
}
