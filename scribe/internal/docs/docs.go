package docs

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/julieqiu/exp/scribe/internal/scraper"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

var (
	templates *template.Template
)

var supportedLanguages = []string{
	"cpp",
	"dotnet",
	"go",
	"java",
	"nodejs",
	"php",
	"python",
	"ruby",
	"rust",
}

var languageTitles = map[string]string{
	"cpp":    "C++",
	"dotnet": ".NET",
	"go":     "Go",
	"java":   "Java",
	"nodejs": "Node.js",
	"php":    "PHP",
	"python": "Python",
	"ruby":   "Ruby",
	"rust":   "Rust",
}

type LanguageInfo struct {
	Code string
	Name string
}

type Server struct {
	librariesCache map[string][]scraper.Library
}

// Run creates and executes the docs server command.
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:  "docs",
		Usage: "run a local HTTP server to browse scraped documentation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
				Usage: "port to run the server on",
			},
		},
		Action: run,
	}

	return cmd.Run(ctx, args)
}

func run(ctx context.Context, cmd *cli.Command) error {
	port := cmd.String("port")

	server := &Server{
		librariesCache: make(map[string][]scraper.Library),
	}

	// Load templates
	var err error
	templates, err = template.ParseGlob("static/templates/*.html")
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Load all YAML files into cache
	if err := server.loadLibraries(); err != nil {
		return fmt.Errorf("failed to load libraries: %w", err)
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", server.handleRoot)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	addr := ":" + port
	fmt.Printf("Starting server on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) loadLibraries() error {
	for _, lang := range supportedLanguages {
		filePath := filepath.Join("testdata", "reference", lang+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		var libraries []scraper.Library
		if err := yaml.Unmarshal(data, &libraries); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", filePath, err)
		}

		s.librariesCache[lang] = libraries
	}

	return nil
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")

	// Root path - list languages
	if path == "" {
		s.handleLanguageList(w, r)
		return
	}

	// Split path into parts
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 1 {
		// /{language} - show table of contents
		s.handleLanguageTOC(w, r, parts[0])
		return
	}

	if len(parts) == 2 {
		// /{language}/{package} - redirect to package docs
		// Package name is everything after the language
		s.handlePackageRedirect(w, r, parts[0], parts[1])
		return
	}

	http.NotFound(w, r)
}

func (s *Server) handleLanguageList(w http.ResponseWriter, r *http.Request) {
	var languages []LanguageInfo
	for _, lang := range supportedLanguages {
		languages = append(languages, LanguageInfo{
			Code: lang,
			Name: languageTitles[lang],
		})
	}
	templates.ExecuteTemplate(w, "languages.html", languages)
}

func (s *Server) handleLanguageTOC(w http.ResponseWriter, r *http.Request, language string) {
	libraries, ok := s.librariesCache[language]
	if !ok {
		http.NotFound(w, r)
		return
	}

	languageTitle, ok := languageTitles[language]
	if !ok {
		http.NotFound(w, r)
		return
	}

	sort.Slice(libraries, func(i, j int) bool {
		return libraries[i].Name < libraries[j].Name
	})

	// Generate original docs URL
	originalDocsURL := fmt.Sprintf("https://cloud.google.com/%s/docs/reference", language)

	data := struct {
		LanguageTitle   string
		Language        string
		OriginalDocsURL string
		Libraries       []scraper.Library
	}{
		LanguageTitle:   languageTitle,
		Language:        language,
		OriginalDocsURL: originalDocsURL,
		Libraries:       libraries,
	}

	templates.ExecuteTemplate(w, "toc.html", data)
}

func (s *Server) handlePackageRedirect(w http.ResponseWriter, r *http.Request, language, packageName string) {
	libraries, ok := s.librariesCache[language]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Find the package and redirect to its link
	for _, lib := range libraries {
		for _, pkg := range lib.Packages {
			if pkg.Name == packageName {
				var redirectURL string
				if language == "go" {
					// For Go, redirect to pkg.go.dev
					redirectURL = "https://pkg.go.dev/" + pkg.Name
				} else if language == "dotnet" && strings.Contains(pkg.Link, "googleapis.dev") {
					// For .NET packages on googleapis.dev, add the /api/{packageName}.html suffix
					redirectURL = fmt.Sprintf("%s/api/%s.html", pkg.Link, pkg.Name)
				} else {
					// For other languages, use the package link
					redirectURL = pkg.Link
				}
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}
		}
	}

	http.NotFound(w, r)
}
