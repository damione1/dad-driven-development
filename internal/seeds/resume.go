package seeds

import (
	"encoding/json"
	"os"
	"time"
)

// Resume represents the jsoncv resume schema
type Resume struct {
	Basics      Basics       `json:"basics"`
	Work        []Work       `json:"work"`
	Volunteer   []Volunteer  `json:"volunteer"`
	Education   []Education  `json:"education"`
	Awards      []Award      `json:"awards"`
	Certificates []Certificate `json:"certificates"`
	Publications []Publication `json:"publications"`
	Skills      []Skill      `json:"skills"`
	Languages   []Language   `json:"languages"`
	Interests   []Interest   `json:"interests"`
	References  []Reference  `json:"references"`
	Projects    []Project    `json:"projects"`
	SideProjects []Project   `json:"sideProjects,omitempty"`
	Meta        Meta         `json:"meta"`
}

// Basics represents the basics section
type Basics struct {
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Image    string    `json:"image"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	URL      string    `json:"url"`
	Summary  string    `json:"summary"`
	Location Location  `json:"location"`
	Profiles []Profile `json:"profiles"`
}

// Location represents location information
type Location struct {
	Address     string `json:"address"`
	PostalCode  string `json:"postalCode"`
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
}

// Profile represents a social profile
type Profile struct {
	Network  string `json:"network"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

// Work represents work experience
type Work struct {
	Name       string   `json:"name"`
	Position   string   `json:"position"`
	Location   string   `json:"location"`
	StartDate  string   `json:"startDate"` // YYYY-MM format
	EndDate    string   `json:"endDate"`   // YYYY-MM format or empty
	URL        string   `json:"url"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

// Volunteer represents volunteer experience
type Volunteer struct {
	Organization string   `json:"organization"`
	Position     string   `json:"position"`
	URL          string   `json:"url"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Summary      string   `json:"summary"`
	Highlights   []string `json:"highlights"`
}

// Education represents education history
type Education struct {
	Institution string   `json:"institution"`
	Area        string   `json:"area"`
	StudyType   string   `json:"studyType"`
	StartDate   string   `json:"startDate"` // YYYY-MM format
	EndDate     string   `json:"endDate"`    // YYYY-MM format or empty
	Score       string   `json:"score"`
	Courses     []string `json:"courses"`
}

// Award represents an award
type Award struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Awarder string `json:"awarder"`
	Summary string `json:"summary"`
}

// Certificate represents a certificate
type Certificate struct {
	Name   string `json:"name"`
	Date   string `json:"date"`
	Issuer string `json:"issuer"`
	URL    string `json:"url"`
}

// Publication represents a publication
type Publication struct {
	Name        string `json:"name"`
	Publisher   string `json:"publisher"`
	ReleaseDate string `json:"releaseDate"`
	URL         string `json:"url"`
	Summary     string `json:"summary"`
}

// Skill represents a skill category
type Skill struct {
	Name     string   `json:"name"`
	Level    string   `json:"level"`
	Keywords []string `json:"keywords"`
}

// Language represents a language
type Language struct {
	Language string `json:"language"`
	Fluency  string `json:"fluency"`
}

// Interest represents an interest
type Interest struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

// Reference represents a reference
type Reference struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

// Project represents a project
type Project struct {
	Name        string   `json:"name"`
	StartDate   string   `json:"startDate"` // YYYY-MM format
	EndDate     string   `json:"endDate"`   // YYYY-MM format or empty
	Description string   `json:"description"`
	Highlights  []string `json:"highlights"`
	Keywords    []string `json:"keywords"`
	Type        string   `json:"type"`
	URL         string   `json:"url"`
	Roles       []string `json:"roles"`
}

// Meta represents metadata
type Meta struct {
	Version      string `json:"version"`
	LastModified string `json:"lastModified"`
	Name         string `json:"name,omitempty"` // jsoncv extension
}

// LoadResume loads and parses a resume JSON file
func LoadResume(filePath string) (*Resume, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var resume Resume
	if err := json.Unmarshal(data, &resume); err != nil {
		return nil, err
	}

	return &resume, nil
}

// ParseDate parses a YYYY-MM date string and returns a time.Time
// If the date is empty, returns zero time
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	// Append "-01" to make it a full date for parsing
	return time.Parse("2006-01-02", dateStr+"-01")
}

// FormatLocation formats location fields into a string
func FormatLocation(loc Location) string {
	parts := []string{}
	if loc.City != "" {
		parts = append(parts, loc.City)
	}
	if loc.Region != "" {
		parts = append(parts, loc.Region)
	}
	if loc.CountryCode != "" {
		parts = append(parts, loc.CountryCode)
	}
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}
