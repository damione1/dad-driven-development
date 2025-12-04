package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Resume struct {
	Basics struct {
		Name     string `json:"name"`
		Label    string `json:"label"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		URL      string `json:"url"`
		Summary  string `json:"summary"`
		Profiles []struct {
			Network  string `json:"network"`
			Username string `json:"username"`
			URL      string `json:"url"`
		} `json:"profiles"`
	} `json:"basics"`
	Work []struct {
		Name       string   `json:"name"`
		Position   string   `json:"position"`
		Location   string   `json:"location"`
		StartDate  string   `json:"startDate"`
		EndDate    string   `json:"endDate,omitempty"`
		URL        string   `json:"url"`
		Summary    string   `json:"summary"`
		Highlights []string `json:"highlights"`
	} `json:"work"`
	Education []struct {
		Institution string   `json:"institution"`
		Area        string   `json:"area"`
		StudyType   string   `json:"studyType"`
		StartDate   string   `json:"startDate"`
		EndDate     string   `json:"endDate"`
		Courses     []string `json:"courses"`
	} `json:"education"`
	Certificates []struct {
		Name   string `json:"name"`
		Date   string `json:"date"`
		Issuer string `json:"issuer"`
		URL    string `json:"url"`
	} `json:"certificates"`
	Skills []struct {
		Name     string   `json:"name"`
		Level    string   `json:"level"`
		Keywords []string `json:"keywords"`
	} `json:"skills"`
	Projects []struct {
		Name        string   `json:"name"`
		StartDate   string   `json:"startDate"`
		URL         string   `json:"url"`
		Description string   `json:"description"`
		Highlights  []string `json:"highlights"`
		Keywords    []string `json:"keywords"`
		Type        string   `json:"type"`
		Roles       []string `json:"roles"`
	} `json:"projects"`
}

func main() {
	data, err := os.ReadFile("old/.claude/resume.json")
	if err != nil {
		fmt.Printf("Error reading resume.json: %v\n", err)
		return
	}

	var resume Resume
	if err := json.Unmarshal(data, &resume); err != nil {
		fmt.Printf("Error parsing resume.json: %v\n", err)
		return
	}

	createAboutPage(&resume)
	for i := range resume.Work {
		createJob(&resume, i)
	}
	for i := range resume.Projects {
		createProject(&resume, i)
	}
	for i := range resume.Certificates {
		createCertification(&resume, i)
	}
	createResumeIndex(&resume)

	fmt.Println("✓ Migration completed successfully!")
}

func createAboutPage(resume *Resume) {
	content := fmt.Sprintf(`---
title: "%s"
---

# About Me

%s

## Skills

`, resume.Basics.Name, resume.Basics.Summary)

	for _, skill := range resume.Skills {
		content += fmt.Sprintf("### %s\n\n", skill.Name)
		for _, keyword := range skill.Keywords {
			content += fmt.Sprintf("- %s\n", keyword)
		}
		content += "\n"
	}

	os.WriteFile("content/en/_index.md", []byte(content), 0644)
	fmt.Println("✓ Created about page")
}

func createJob(resume *Resume, idx int) {
	job := resume.Work[idx]
	slug := strings.ToLower(strings.ReplaceAll(job.Name, " ", "-"))
	slug = strings.ReplaceAll(slug, ",", "")

	stack := extractStack(job.Summary, job.Highlights, resume.Skills)
	endDate := job.EndDate
	if endDate == "" {
		endDate = "Present"
	}

	content := fmt.Sprintf(`---
title: "%s"
company: "%s"
position: "%s"
location: "%s"
startDate: "%s"
endDate: "%s"
url: "%s"
stack: %s
draft: false
weight: %s
---

%s

## Highlights

`, job.Position, job.Name, job.Position, job.Location, job.StartDate, endDate, job.URL, formatStack(stack), getWeight(job.StartDate), job.Summary)

	for _, highlight := range job.Highlights {
		content += fmt.Sprintf("- %s\n", highlight)
	}

	filename := fmt.Sprintf("content/en/resume/jobs/%s.md", slug)
	os.WriteFile(filename, []byte(content), 0644)
	fmt.Printf("✓ Created job: %s\n", job.Name)
}

func createProject(resume *Resume, idx int) {
	project := resume.Projects[idx]
	slug := strings.ToLower(strings.ReplaceAll(project.Name, " ", "-"))

	content := fmt.Sprintf(`---
title: "%s"
date: %s
draft: false
stack: %s
description: "%s"
url: "%s"
startDate: "%s"
type: "%s"
roles: %s
---

%s

## Highlights

`, project.Name, formatDate(project.StartDate), formatStack(project.Keywords), project.Description, project.URL, project.StartDate, project.Type, formatRoles(project.Roles), project.Description)

	for _, highlight := range project.Highlights {
		content += fmt.Sprintf("- %s\n", highlight)
	}

	filename := fmt.Sprintf("content/en/projects/%s.md", slug)
	os.WriteFile(filename, []byte(content), 0644)
	fmt.Printf("✓ Created project: %s\n", project.Name)
}

func createCertification(resume *Resume, idx int) {
	cert := resume.Certificates[idx]
	slug := strings.ToLower(strings.ReplaceAll(cert.Name, " ", "-"))

	content := fmt.Sprintf(`---
title: "%s"
issuer: "%s"
date: "%s"
url: "%s"
stack: ["%s"]
draft: false
---

Certification in %s issued by %s.
`, cert.Name, cert.Issuer, cert.Date, cert.URL, cert.Name, cert.Name, cert.Issuer)

	filename := fmt.Sprintf("content/en/resume/certifications/%s.md", slug)
	os.WriteFile(filename, []byte(content), 0644)
	fmt.Printf("✓ Created certification: %s\n", cert.Name)
}

func createResumeIndex(resume *Resume) {
	content := `---
title: "Resume"
---

# Professional Experience

See my work history, projects, certifications, and education below.

## Education

`

	for _, edu := range resume.Education {
		content += fmt.Sprintf("### %s\n\n", edu.Institution)
		content += fmt.Sprintf("**%s** - %s\n\n", edu.StudyType, edu.Area)
		content += fmt.Sprintf("*%s to %s*\n\n", formatDate(edu.StartDate), formatDate(edu.EndDate))

		if len(edu.Courses) > 0 {
			content += "**Courses:**\n\n"
			for _, course := range edu.Courses {
				content += fmt.Sprintf("- %s\n", course)
			}
			content += "\n"
		}
	}

	os.WriteFile("content/en/resume/_index.md", []byte(content), 0644)
	fmt.Println("✓ Created resume index page")
}

func extractStack(summary string, highlights []string, skills []struct {
	Name     string   `json:"name"`
	Level    string   `json:"level"`
	Keywords []string `json:"keywords"`
}) []string {
	keywordMap := make(map[string]bool)
	text := strings.ToLower(summary + " " + strings.Join(highlights, " "))

	for _, skill := range skills {
		for _, keyword := range skill.Keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				keywordMap[keyword] = true
			}
		}
	}

	keywords := make([]string, 0, len(keywordMap))
	for k := range keywordMap {
		keywords = append(keywords, k)
	}

	return keywords
}

func formatStack(stack []string) string {
	if len(stack) == 0 {
		return "[]"
	}

	quoted := make([]string, len(stack))
	for i, s := range stack {
		quoted[i] = fmt.Sprintf(`"%s"`, s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func formatRoles(roles []string) string {
	if len(roles) == 0 {
		return "[]"
	}

	quoted := make([]string, len(roles))
	for i, r := range roles {
		quoted[i] = fmt.Sprintf(`"%s"`, r)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return time.Now().Format("2006-01-02")
	}

	t, err := time.Parse("2006-01", dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("2006-01-02")
}

func getWeight(startDate string) string {
	t, err := time.Parse("2006-01", startDate)
	if err != nil {
		return "100"
	}

	monthsSince := int(time.Since(t).Hours() / 24 / 30)
	return fmt.Sprintf("%d", monthsSince)
}
