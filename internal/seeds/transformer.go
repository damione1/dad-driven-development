package seeds

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Transformer handles transformation of resume data to PocketBase records
type Transformer struct {
	app           core.App
	stackNameToID map[string]string // Maps stack item name to record ID
}

// NewTransformer creates a new transformer
func NewTransformer(app core.App) *Transformer {
	return &Transformer{
		app:           app,
		stackNameToID: make(map[string]string),
	}
}

// SetStackMap sets the stack name to ID mapping
func (t *Transformer) SetStackMap(m map[string]string) {
	t.stackNameToID = m
}

// TransformProfile transforms basics to a profile record
func (t *Transformer) TransformProfile(basics Basics) (*core.Record, error) {
	collection, err := t.app.FindCollectionByNameOrId("profile")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)

	record.Set("full_name", basics.Name)
	record.Set("title", basics.Label)
	record.Set("bio", basics.Summary)
	record.Set("email", basics.Email)
	record.Set("location", FormatLocation(basics.Location))

	// Transform profiles to social_links JSON
	socialLinks := make(map[string]string)
	for _, profile := range basics.Profiles {
		network := strings.ToLower(profile.Network)
		socialLinks[network] = profile.URL
	}
	if len(socialLinks) > 0 {
		record.Set("social_links", socialLinks)
	}

	return record, nil
}

// TransformStackItems extracts all unique keywords from skills and creates stack items
func (t *Transformer) TransformStackItems(skills []Skill) ([]*core.Record, error) {
	collection, err := t.app.FindCollectionByNameOrId("stack")
	if err != nil {
		return nil, err
	}

	// Extract all unique keywords
	keywordSet := make(map[string]bool)
	for _, skill := range skills {
		for _, keyword := range skill.Keywords {
			keywordSet[keyword] = true
		}
	}

	var records []*core.Record
	for keyword := range keywordSet {
		record := core.NewRecord(collection)
		record.Set("name", keyword)
		record.Set("category", inferCategory(keyword))
		record.Set("description", "")
		record.Set("featured", false)
		records = append(records, record)
	}

	return records, nil
}

// TransformExperiences transforms work experience to experience records
func (t *Transformer) TransformExperiences(work []Work) ([]*core.Record, error) {
	collection, err := t.app.FindCollectionByNameOrId("experiences")
	if err != nil {
		return nil, err
	}

	var records []*core.Record
	for _, w := range work {
		record := core.NewRecord(collection)

		record.Set("company", w.Name)
		record.Set("position", w.Position)
		record.Set("location", w.Location)
		record.Set("description", w.Summary)

		// Parse dates
		startDate, err := ParseDate(w.StartDate)
		if err != nil {
			log.Printf("Warning: failed to parse start date %s: %v", w.StartDate, err)
		} else if !startDate.IsZero() {
			record.Set("start_date", startDate.Format("2006-01-02"))
		}

		if w.EndDate != "" {
			endDate, err := ParseDate(w.EndDate)
			if err != nil {
				log.Printf("Warning: failed to parse end date %s: %v", w.EndDate, err)
			} else if !endDate.IsZero() {
				record.Set("end_date", endDate.Format("2006-01-02"))
			}
		}

		// Set highlights
		if len(w.Highlights) > 0 {
			record.Set("highlights", w.Highlights)
		}

		// Match stack items from position, summary, and highlights
		stackIDs := t.matchStackItems(w.Position + " " + w.Summary + " " + strings.Join(w.Highlights, " "))
		if len(stackIDs) > 0 {
			record.Set("stack_items", stackIDs)
		}

		records = append(records, record)
	}

	return records, nil
}

// TransformEducation transforms education to education records
func (t *Transformer) TransformEducation(edu []Education) ([]*core.Record, error) {
	collection, err := t.app.FindCollectionByNameOrId("education")
	if err != nil {
		return nil, err
	}

	var records []*core.Record
	for _, e := range edu {
		record := core.NewRecord(collection)

		record.Set("institution", e.Institution)
		record.Set("degree", e.StudyType)
		record.Set("field", e.Area)

		// Format description from courses
		if len(e.Courses) > 0 {
			record.Set("description", strings.Join(e.Courses, ", "))
		}

		// Parse dates
		startDate, err := ParseDate(e.StartDate)
		if err != nil {
			log.Printf("Warning: failed to parse start date %s: %v", e.StartDate, err)
		} else if !startDate.IsZero() {
			record.Set("start_date", startDate.Format("2006-01-02"))
		}

		if e.EndDate != "" {
			endDate, err := ParseDate(e.EndDate)
			if err != nil {
				log.Printf("Warning: failed to parse end date %s: %v", e.EndDate, err)
			} else if !endDate.IsZero() {
				record.Set("end_date", endDate.Format("2006-01-02"))
			}
		}

		// Match stack items from area and courses
		text := e.Area + " " + strings.Join(e.Courses, " ")
		stackIDs := t.matchStackItems(text)
		if len(stackIDs) > 0 {
			record.Set("stack_items", stackIDs)
		}

		records = append(records, record)
	}

	return records, nil
}

// TransformProjects transforms projects to project records
func (t *Transformer) TransformProjects(projects []Project) ([]*core.Record, error) {
	collection, err := t.app.FindCollectionByNameOrId("projects")
	if err != nil {
		return nil, err
	}

	var records []*core.Record
	for i, p := range projects {
		record := core.NewRecord(collection)

		record.Set("name", p.Name)
		record.Set("slug", generateSlug(p.Name))
		record.Set("status", "active")

		// Truncate description to 500 chars
		description := p.Description
		if len(description) > 500 {
			description = description[:497] + "..."
		}
		record.Set("description", description)

		// Format content from highlights
		var content strings.Builder
		if len(p.Highlights) > 0 {
			for _, highlight := range p.Highlights {
				content.WriteString("- ")
				content.WriteString(highlight)
				content.WriteString("\n")
			}
		}
		if content.Len() > 0 {
			record.Set("content", content.String())
		}

		// Extract GitHub URL and live URL
		if strings.Contains(p.URL, "github.com") {
			record.Set("github_url", p.URL)
		} else if p.URL != "" {
			record.Set("live_url", p.URL)
		}

		// Set featured for first 2 projects
		record.Set("featured", i < 2)

		// Parse published_at from startDate or use current date
		if p.StartDate != "" {
			pubDate, err := ParseDate(p.StartDate)
			if err == nil && !pubDate.IsZero() {
				record.Set("published_at", pubDate.Format("2006-01-02"))
			} else {
				record.Set("published_at", time.Now().Format("2006-01-02"))
			}
		} else {
			record.Set("published_at", time.Now().Format("2006-01-02"))
		}

		// Match stack items from keywords
		if len(p.Keywords) > 0 {
			stackIDs := t.matchStackItemsFromKeywords(p.Keywords)
			if len(stackIDs) > 0 {
				record.Set("stack_items", stackIDs)
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// matchStackItems matches keywords in text to stack items (exact name match, case-sensitive)
func (t *Transformer) matchStackItems(text string) []string {
	var stackIDs []string
	words := extractWords(text)

	for _, word := range words {
		if id, exists := t.stackNameToID[word]; exists {
			// Avoid duplicates
			found := false
			for _, existingID := range stackIDs {
				if existingID == id {
					found = true
					break
				}
			}
			if !found {
				stackIDs = append(stackIDs, id)
			}
		}
	}

	return stackIDs
}

// matchStackItemsFromKeywords matches keywords array to stack items
func (t *Transformer) matchStackItemsFromKeywords(keywords []string) []string {
	var stackIDs []string

	for _, keyword := range keywords {
		if id, exists := t.stackNameToID[keyword]; exists {
			// Avoid duplicates
			found := false
			for _, existingID := range stackIDs {
				if existingID == id {
					found = true
					break
				}
			}
			if !found {
				stackIDs = append(stackIDs, id)
			}
		}
	}

	return stackIDs
}

// extractWords extracts potential technology names from text
func extractWords(text string) []string {
	// Simple word extraction - split on whitespace and punctuation
	re := regexp.MustCompile(`[A-Za-z0-9]+`)
	matches := re.FindAllString(text, -1)
	return matches
}

// generateSlug generates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces and special chars with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

// inferCategory infers the stack category from a keyword name
func inferCategory(keyword string) string {
	keywordLower := strings.ToLower(keyword)

	// Backend
	backendKeywords := []string{"go", "php", "python", "java", "node", "grpc", "api", "rest", "graphql", "microservices"}
	for _, bk := range backendKeywords {
		if strings.Contains(keywordLower, bk) {
			return "backend"
		}
	}

	// Frontend
	frontendKeywords := []string{"react", "vue", "angular", "javascript", "typescript", "html", "css", "htmx", "alpine", "tailwind"}
	for _, fk := range frontendKeywords {
		if strings.Contains(keywordLower, fk) {
			return "frontend"
		}
	}

	// Database
	databaseKeywords := []string{"postgresql", "postgres", "mysql", "sqlite", "redis", "mongodb", "sql", "database", "db"}
	for _, dk := range databaseKeywords {
		if strings.Contains(keywordLower, dk) {
			return "database"
		}
	}

	// DevOps
	devopsKeywords := []string{"docker", "kubernetes", "k8s", "terraform", "ci/cd", "cicd", "gcp", "aws", "azure", "github", "gitlab", "jenkins", "ansible", "puppet", "chef"}
	for _, dok := range devopsKeywords {
		if strings.Contains(keywordLower, dok) {
			return "devops"
		}
	}

	// Tools
	toolsKeywords := []string{"git", "vscode", "vim", "neovim", "ide", "editor"}
	for _, tk := range toolsKeywords {
		if strings.Contains(keywordLower, tk) {
			return "tools"
		}
	}

	// Default to tools if no match
	return "tools"
}
