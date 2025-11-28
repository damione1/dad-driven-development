package services

import (
	"errors"

	"github.com/pocketbase/pocketbase/core"
)

type ContentManager struct {
	app core.App
}

func NewContentManager(app core.App) *ContentManager {
	return &ContentManager{app: app}
}

// Profile
func (cm *ContentManager) GetProfile() (*core.Record, error) {
	records, err := cm.app.FindRecordsByFilter("profile", "", "", 1, 0)
	if err != nil || len(records) == 0 {
		return nil, errors.New("profile not found")
	}
	return records[0], nil
}

// Stack Items
func (cm *ContentManager) GetAllStackItems() ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"stack",
		"",
		"category, sort_order, name",
		100,
		0,
	)
}

func (cm *ContentManager) GetFeaturedStackItems() ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"stack",
		"featured = true",
		"category, sort_order, name",
		50,
		0,
	)
}

func (cm *ContentManager) GetStackItemsByIDs(ids []string) ([]*core.Record, error) {
	if len(ids) == 0 {
		return []*core.Record{}, nil
	}
	return cm.app.FindRecordsByIds("stack", ids)
}

// Projects
func (cm *ContentManager) GetAllProjects(limit, offset int) ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"projects",
		"status = 'active'",
		"-published_at, -sort_order",
		limit,
		offset,
	)
}

func (cm *ContentManager) GetProjectBySlug(slug string) (*core.Record, error) {
	records, err := cm.app.FindRecordsByFilter(
		"projects",
		"slug = {:slug}",
		"",
		1,
		0,
		map[string]any{"slug": slug},
	)
	if err != nil || len(records) == 0 {
		return nil, errors.New("project not found")
	}
	return records[0], nil
}

func (cm *ContentManager) GetFeaturedProjects() ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"projects",
		"featured = true && status = 'active'",
		"-sort_order",
		6,
		0,
	)
}

// Experiences
func (cm *ContentManager) GetAllExperiences() ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"experiences",
		"",
		"-start_date, -sort_order",
		100,
		0,
	)
}

// Education
func (cm *ContentManager) GetAllEducation() ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"education",
		"",
		"-start_date, -sort_order",
		100,
		0,
	)
}

// Blog Posts
func (cm *ContentManager) GetPublishedPosts(limit, offset int) ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"posts",
		"status = 'published'",
		"-published_at",
		limit,
		offset,
	)
}

func (cm *ContentManager) GetPostBySlug(slug string) (*core.Record, error) {
	records, err := cm.app.FindRecordsByFilter(
		"posts",
		"slug = {:slug} && status = 'published'",
		"",
		1,
		0,
		map[string]any{"slug": slug},
	)
	if err != nil || len(records) == 0 {
		return nil, errors.New("post not found")
	}
	return records[0], nil
}

func (cm *ContentManager) GetRecentPosts(limit int) ([]*core.Record, error) {
	return cm.app.FindRecordsByFilter(
		"posts",
		"status = 'published'",
		"-published_at",
		limit,
		0,
	)
}
