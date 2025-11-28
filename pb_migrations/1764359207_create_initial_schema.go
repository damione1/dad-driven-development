package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create stack collection
		stack := core.NewBaseCollection("stack")
		stack.ListRule = nil // Public read - set via API rules if needed
		stack.ViewRule = nil
		stack.CreateRule = nil // Admin only
		stack.UpdateRule = nil
		stack.DeleteRule = nil

		stack.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      100,
		})

		stack.Fields.Add(&core.SelectField{
			Name:      "category",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"backend", "frontend", "database", "devops", "tools"},
		})

		stack.Fields.Add(&core.EditorField{
			Name:     "description",
			Required: false,
			MaxSize:  2048,
		})

		stack.Fields.Add(&core.URLField{
			Name:     "icon_url",
			Required: false,
		})

		stack.Fields.Add(&core.URLField{
			Name:     "url",
			Required: false,
		})

		stack.Fields.Add(&core.NumberField{
			Name:     "sort_order",
			Required: false,
		})

		stack.Fields.Add(&core.BoolField{
			Name:     "featured",
			Required: false,
		})

		stack.Indexes = []string{
			"CREATE INDEX idx_stack_name ON stack(name)",
			"CREATE INDEX idx_stack_category ON stack(category)",
			"CREATE INDEX idx_stack_featured ON stack(featured)",
		}

		if err := app.Save(stack); err != nil {
			return err
		}

		// Create experiences collection
		experiences := core.NewBaseCollection("experiences")
		experiences.ListRule = nil
		experiences.ViewRule = nil
		experiences.CreateRule = nil
		experiences.UpdateRule = nil
		experiences.DeleteRule = nil

		experiences.Fields.Add(&core.TextField{
			Name:     "company",
			Required: true,
			Max:      200,
		})

		experiences.Fields.Add(&core.TextField{
			Name:     "position",
			Required: true,
			Max:      200,
		})

		experiences.Fields.Add(&core.TextField{
			Name:     "location",
			Required: false,
			Max:      200,
		})

		experiences.Fields.Add(&core.DateField{
			Name:     "start_date",
			Required: true,
		})

		experiences.Fields.Add(&core.DateField{
			Name:     "end_date",
			Required: false, // Null means current position
		})

		experiences.Fields.Add(&core.EditorField{
			Name:     "description",
			Required: false,
			MaxSize:  10240,
		})

		experiences.Fields.Add(&core.JSONField{
			Name:     "highlights",
			Required: false,
			MaxSize:  5120,
		})

		experiences.Fields.Add(&core.RelationField{
			Name:          "stack_items",
			Required:      false,
			CollectionId:  stack.Id,
			CascadeDelete: false,
			MaxSelect:     50,
		})

		experiences.Fields.Add(&core.NumberField{
			Name:     "sort_order",
			Required: false,
		})

		experiences.Indexes = []string{
			"CREATE INDEX idx_experiences_dates ON experiences(start_date, end_date)",
		}

		if err := app.Save(experiences); err != nil {
			return err
		}

		// Create education collection
		education := core.NewBaseCollection("education")
		education.ListRule = nil
		education.ViewRule = nil
		education.CreateRule = nil
		education.UpdateRule = nil
		education.DeleteRule = nil

		education.Fields.Add(&core.TextField{
			Name:     "institution",
			Required: true,
			Max:      200,
		})

		education.Fields.Add(&core.TextField{
			Name:     "degree",
			Required: true,
			Max:      200,
		})

		education.Fields.Add(&core.TextField{
			Name:     "field",
			Required: true,
			Max:      200,
		})

		education.Fields.Add(&core.DateField{
			Name:     "start_date",
			Required: true,
		})

		education.Fields.Add(&core.DateField{
			Name:     "end_date",
			Required: false,
		})

		education.Fields.Add(&core.EditorField{
			Name:     "description",
			Required: false,
			MaxSize:  5120,
		})

		education.Fields.Add(&core.RelationField{
			Name:          "stack_items",
			Required:      false,
			CollectionId:  stack.Id,
			CascadeDelete: false,
			MaxSelect:     50,
			
		})

		education.Fields.Add(&core.NumberField{
			Name:     "sort_order",
			Required: false,
		})

		if err := app.Save(education); err != nil {
			return err
		}

		// Create projects collection
		projects := core.NewBaseCollection("projects")
		projects.ListRule = nil
		projects.ViewRule = nil
		projects.CreateRule = nil
		projects.UpdateRule = nil
		projects.DeleteRule = nil

		projects.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      200,
		})

		projects.Fields.Add(&core.TextField{
			Name:     "slug",
			Required: true,
			Max:      200,
			Pattern:  `^[a-z0-9-]+$`,
		})

		projects.Fields.Add(&core.TextField{
			Name:     "description",
			Required: true,
			Max:      500,
		})

		projects.Fields.Add(&core.EditorField{
			Name:     "content",
			Required: false,
			MaxSize:  20480,
		})

		projects.Fields.Add(&core.FileField{
			Name:        "thumbnail",
			Required:    false,
			MaxSelect:   1,
			MaxSize:     5242880, // 5MB
			MimeTypes:   []string{"image/jpeg", "image/png", "image/webp"},
			Thumbs:      []string{"100x100", "400x300", "800x600"},
			Protected:   false,
		})

		projects.Fields.Add(&core.FileField{
			Name:        "images",
			Required:    false,
			MaxSelect:   10,
			MaxSize:     5242880,
			MimeTypes:   []string{"image/jpeg", "image/png", "image/webp"},
			Thumbs:      []string{"400x300", "800x600"},
			Protected:   false,
		})

		projects.Fields.Add(&core.RelationField{
			Name:          "stack_items",
			Required:      false,
			CollectionId:  stack.Id,
			CascadeDelete: false,
			MaxSelect:     50,
			
		})

		projects.Fields.Add(&core.URLField{
			Name:     "github_url",
			Required: false,
		})

		projects.Fields.Add(&core.URLField{
			Name:     "live_url",
			Required: false,
		})

		projects.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"active", "archived", "in_progress"},
		})

		projects.Fields.Add(&core.BoolField{
			Name:     "featured",
			Required: false,
		})

		projects.Fields.Add(&core.DateField{
			Name:     "published_at",
			Required: true,
		})

		projects.Fields.Add(&core.NumberField{
			Name:     "sort_order",
			Required: false,
		})

		projects.Indexes = []string{
			"CREATE UNIQUE INDEX idx_projects_slug ON projects(slug)",
			"CREATE INDEX idx_projects_status ON projects(status)",
			"CREATE INDEX idx_projects_featured ON projects(featured)",
			"CREATE INDEX idx_projects_published ON projects(published_at)",
		}

		if err := app.Save(projects); err != nil {
			return err
		}

		// Create posts collection
		posts := core.NewBaseCollection("posts")
		posts.ListRule = nil
		posts.ViewRule = nil
		posts.CreateRule = nil
		posts.UpdateRule = nil
		posts.DeleteRule = nil

		posts.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Max:      200,
		})

		posts.Fields.Add(&core.TextField{
			Name:     "slug",
			Required: true,
			Max:      200,
			Pattern:  `^[a-z0-9-]+$`,
		})

		posts.Fields.Add(&core.TextField{
			Name:     "excerpt",
			Required: true,
			Max:      500,
		})

		posts.Fields.Add(&core.EditorField{
			Name:     "content",
			Required: true,
			MaxSize:  102400, // 100KB for blog posts
		})

		posts.Fields.Add(&core.FileField{
			Name:        "thumbnail",
			Required:    false,
			MaxSelect:   1,
			MaxSize:     5242880,
			MimeTypes:   []string{"image/jpeg", "image/png", "image/webp"},
			Thumbs:      []string{"400x300", "800x600"},
			Protected:   false,
		})

		posts.Fields.Add(&core.JSONField{
			Name:     "tags",
			Required: false,
			MaxSize:  2048,
		})

		posts.Fields.Add(&core.RelationField{
			Name:          "stack_items",
			Required:      false,
			CollectionId:  stack.Id,
			CascadeDelete: false,
			MaxSelect:     50,
			
		})

		posts.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"draft", "published"},
		})

		posts.Fields.Add(&core.DateField{
			Name:     "published_at",
			Required: false,
		})

		posts.Fields.Add(&core.NumberField{
			Name:     "view_count",
			Required: false,
		})

		posts.Indexes = []string{
			"CREATE UNIQUE INDEX idx_posts_slug ON posts(slug)",
			"CREATE INDEX idx_posts_status ON posts(status)",
			"CREATE INDEX idx_posts_published ON posts(published_at)",
		}

		if err := app.Save(posts); err != nil {
			return err
		}

		// Create profile collection (single record)
		profile := core.NewBaseCollection("profile")
		profile.ListRule = nil
		profile.ViewRule = nil
		profile.CreateRule = nil
		profile.UpdateRule = nil
		profile.DeleteRule = nil

		profile.Fields.Add(&core.TextField{
			Name:     "full_name",
			Required: true,
			Max:      100,
		})

		profile.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Max:      200,
		})

		profile.Fields.Add(&core.EditorField{
			Name:     "bio",
			Required: true,
			MaxSize:  10240,
		})

		profile.Fields.Add(&core.FileField{
			Name:        "avatar",
			Required:    false,
			MaxSelect:   1,
			MaxSize:     2097152, // 2MB
			MimeTypes:   []string{"image/jpeg", "image/png", "image/webp"},
			Thumbs:      []string{"100x100", "200x200", "400x400"},
			Protected:   false,
		})

		profile.Fields.Add(&core.FileField{
			Name:        "resume",
			Required:    false,
			MaxSelect:   1,
			MaxSize:     10485760, // 10MB
			MimeTypes:   []string{"application/pdf"},
			Protected:   false,
		})

		profile.Fields.Add(&core.EmailField{
			Name:     "email",
			Required: true,
		})

		profile.Fields.Add(&core.TextField{
			Name:     "location",
			Required: false,
			Max:      200,
		})

		profile.Fields.Add(&core.JSONField{
			Name:     "social_links",
			Required: false,
			MaxSize:  2048,
		})

		return app.Save(profile)
	}, func(app core.App) error {
		// Down migration - delete in reverse order
		collections := []string{"profile", "posts", "projects", "education", "experiences", "stack"}

		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err == nil && collection != nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
