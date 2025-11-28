package handlers

import (
	"net/http"

	"github.com/damione1/personal-website/internal/services"
	"github.com/pocketbase/pocketbase/core"
)

type HomeHandler struct {
	contentManager *services.ContentManager
}

func NewHomeHandler(cm *services.ContentManager) *HomeHandler {
	return &HomeHandler{
		contentManager: cm,
	}
}

func (h *HomeHandler) Home(e *core.RequestEvent) error {
	profile, err := h.contentManager.GetProfile()
	if err != nil {
		// For now, return simple HTML. We'll replace with templ templates later
		return e.HTML(http.StatusOK, `
			<html>
			<head><title>Personal Website</title></head>
			<body>
				<h1>Welcome to My Personal Website</h1>
				<p>Profile not yet configured. Please use the admin UI to create your profile.</p>
				<a href="/_/">Go to Admin UI</a>
			</body>
			</html>
		`)
	}

	// For now, return basic HTML. We'll replace with templ templates
	html := `
		<html>
		<head><title>` + profile.GetString("full_name") + `</title></head>
		<body>
			<h1>` + profile.GetString("full_name") + `</h1>
			<p>` + profile.GetString("title") + `</p>
			<nav>
				<a href="/about">About</a> |
				<a href="/experience">Experience</a> |
				<a href="/projects">Projects</a> |
				<a href="/blog">Blog</a> |
				<a href="/stack">Stack</a>
			</nav>
			<p>This is a placeholder. Templ templates coming soon!</p>
		</body>
		</html>
	`

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) About(e *core.RequestEvent) error {
	profile, err := h.contentManager.GetProfile()
	if err != nil {
		return e.String(http.StatusNotFound, "Profile not found")
	}

	experiences, _ := h.contentManager.GetAllExperiences()
	education, _ := h.contentManager.GetAllEducation()

	html := `
		<html>
		<head><title>About - ` + profile.GetString("full_name") + `</title></head>
		<body>
			<h1>About ` + profile.GetString("full_name") + `</h1>
			<p>` + profile.GetString("bio") + `</p>
			<h2>Experience (` + string(rune(len(experiences))) + `)</h2>
			<h2>Education (` + string(rune(len(education))) + `)</h2>
			<a href="/">Home</a>
		</body>
		</html>
	`

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) ExperienceList(e *core.RequestEvent) error {
	experiences, err := h.contentManager.GetAllExperiences()
	if err != nil {
		return e.String(http.StatusInternalServerError, "Error loading experiences")
	}

	html := "<html><head><title>Experience</title></head><body><h1>Experience</h1><ul>"
	for _, exp := range experiences {
		html += "<li>" + exp.GetString("position") + " at " + exp.GetString("company") + "</li>"
	}
	html += "</ul><a href='/'>Home</a></body></html>"

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) ProjectList(e *core.RequestEvent) error {
	projects, err := h.contentManager.GetAllProjects(50, 0)
	if err != nil {
		return e.String(http.StatusInternalServerError, "Error loading projects")
	}

	html := "<html><head><title>Projects</title></head><body><h1>Projects</h1><ul>"
	for _, project := range projects {
		html += "<li><a href='/projects/" + project.GetString("slug") + "'>" + project.GetString("name") + "</a></li>"
	}
	html += "</ul><a href='/'>Home</a></body></html>"

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) ProjectDetail(e *core.RequestEvent) error {
	slug := e.Request.PathValue("slug")
	project, err := h.contentManager.GetProjectBySlug(slug)
	if err != nil {
		return e.Redirect(http.StatusSeeOther, "/projects?error=not_found")
	}

	html := `
		<html>
		<head><title>` + project.GetString("name") + `</title></head>
		<body>
			<h1>` + project.GetString("name") + `</h1>
			<p>` + project.GetString("description") + `</p>
			<div>` + project.GetString("content") + `</div>
			<a href='/projects'>Back to Projects</a>
		</body>
		</html>
	`

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) BlogList(e *core.RequestEvent) error {
	posts, err := h.contentManager.GetPublishedPosts(50, 0)
	if err != nil {
		return e.String(http.StatusInternalServerError, "Error loading blog posts")
	}

	html := "<html><head><title>Blog</title></head><body><h1>Blog</h1><ul>"
	for _, post := range posts {
		html += "<li><a href='/blog/" + post.GetString("slug") + "'>" + post.GetString("title") + "</a></li>"
	}
	html += "</ul><a href='/'>Home</a></body></html>"

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) BlogDetail(e *core.RequestEvent) error {
	slug := e.Request.PathValue("slug")
	post, err := h.contentManager.GetPostBySlug(slug)
	if err != nil {
		return e.Redirect(http.StatusSeeOther, "/blog?error=not_found")
	}

	html := `
		<html>
		<head><title>` + post.GetString("title") + `</title></head>
		<body>
			<h1>` + post.GetString("title") + `</h1>
			<p><em>` + post.GetString("excerpt") + `</em></p>
			<div>` + post.GetString("content") + `</div>
			<a href='/blog'>Back to Blog</a>
		</body>
		</html>
	`

	return e.HTML(http.StatusOK, html)
}

func (h *HomeHandler) StackPage(e *core.RequestEvent) error {
	stackItems, err := h.contentManager.GetAllStackItems()
	if err != nil {
		return e.String(http.StatusInternalServerError, "Error loading stack items")
	}

	html := "<html><head><title>Tech Stack</title></head><body><h1>Tech Stack</h1><ul>"
	for _, item := range stackItems {
		html += "<li><strong>" + item.GetString("name") + "</strong> (" + item.GetString("category") + ")</li>"
	}
	html += "</ul><a href='/'>Home</a></body></html>"

	return e.HTML(http.StatusOK, html)
}
