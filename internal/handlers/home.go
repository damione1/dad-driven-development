package handlers

import (
	"net/http"

	"github.com/damione1/personal-website/internal/services"
	"github.com/damione1/personal-website/web/templates/pages"
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
	profile, _ := h.contentManager.GetProfile()
	featuredProjects, _ := h.contentManager.GetFeaturedProjects()
	recentPosts, _ := h.contentManager.GetRecentPosts(3)

	// Get stack items for each featured project
	projectStacks := make(map[string][]*core.Record)
	for _, project := range featuredProjects {
		stackIDs := project.GetStringSlice("stack_items")
		if len(stackIDs) > 0 {
			stackItems, _ := h.contentManager.GetStackItemsByIDs(stackIDs)
			projectStacks[project.Id] = stackItems
		}
	}

	component := pages.Home(profile, featuredProjects, recentPosts, projectStacks)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) About(e *core.RequestEvent) error {
	profile, _ := h.contentManager.GetProfile()
	experiences, _ := h.contentManager.GetAllExperiences()
	education, _ := h.contentManager.GetAllEducation()

	component := pages.About(profile, len(experiences), len(education))
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) ExperienceList(e *core.RequestEvent) error {
	experiences, _ := h.contentManager.GetAllExperiences()

	// Get stack items for each experience
	experienceStacks := make(map[string][]*core.Record)
	for _, exp := range experiences {
		stackIDs := exp.GetStringSlice("stack_items")
		if len(stackIDs) > 0 {
			stackItems, _ := h.contentManager.GetStackItemsByIDs(stackIDs)
			experienceStacks[exp.Id] = stackItems
		}
	}

	component := pages.Experience(experiences, experienceStacks)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) ProjectList(e *core.RequestEvent) error {
	projects, _ := h.contentManager.GetAllProjects(50, 0)

	// Get stack items for each project
	projectStacks := make(map[string][]*core.Record)
	for _, project := range projects {
		stackIDs := project.GetStringSlice("stack_items")
		if len(stackIDs) > 0 {
			stackItems, _ := h.contentManager.GetStackItemsByIDs(stackIDs)
			projectStacks[project.Id] = stackItems
		}
	}

	component := pages.ProjectsList(projects, projectStacks)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) ProjectDetail(e *core.RequestEvent) error {
	slug := e.Request.PathValue("slug")
	project, err := h.contentManager.GetProjectBySlug(slug)
	if err != nil {
		return e.Redirect(http.StatusSeeOther, "/projects?error=not_found")
	}

	// Get stack items for this project
	var stackItems []*core.Record
	stackIDs := project.GetStringSlice("stack_items")
	if len(stackIDs) > 0 {
		stackItems, _ = h.contentManager.GetStackItemsByIDs(stackIDs)
	}

	component := pages.ProjectDetail(project, stackItems)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) BlogList(e *core.RequestEvent) error {
	posts, _ := h.contentManager.GetPublishedPosts(50, 0)

	component := pages.BlogList(posts)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) BlogDetail(e *core.RequestEvent) error {
	slug := e.Request.PathValue("slug")
	post, err := h.contentManager.GetPostBySlug(slug)
	if err != nil {
		return e.Redirect(http.StatusSeeOther, "/blog?error=not_found")
	}

	// Get stack items for this post
	var stackItems []*core.Record
	stackIDs := post.GetStringSlice("stack_items")
	if len(stackIDs) > 0 {
		stackItems, _ = h.contentManager.GetStackItemsByIDs(stackIDs)
	}

	component := pages.BlogDetail(post, stackItems)
	return component.Render(e.Request.Context(), e.Response)
}

func (h *HomeHandler) StackPage(e *core.RequestEvent) error {
	stackItems, _ := h.contentManager.GetAllStackItems()

	component := pages.Stack(stackItems)
	return component.Render(e.Request.Context(), e.Response)
}
