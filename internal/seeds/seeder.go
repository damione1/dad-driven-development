package seeds

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// Seeder handles the seeding process
type Seeder struct {
	app        core.App
	transformer *Transformer
	resume     *Resume
}

// NewSeeder creates a new seeder
func NewSeeder(app core.App) *Seeder {
	return &Seeder{
		app:        app,
		transformer: NewTransformer(app),
	}
}

// Seed is the main entry point for seeding
func (s *Seeder) Seed(resumePath string, clearExisting bool, dryRun bool) error {
	// Load resume
	log.Printf("Loading resume from: %s", resumePath)
	resume, err := LoadResume(resumePath)
	if err != nil {
		return fmt.Errorf("failed to load resume: %w", err)
	}
	s.resume = resume

	if dryRun {
		log.Println("DRY RUN MODE - No changes will be made")
	}

	// Clear existing data if requested
	if clearExisting && !dryRun {
		log.Println("Clearing existing data...")
		if err := s.clearAllCollections(); err != nil {
			return fmt.Errorf("failed to clear existing data: %w", err)
		}
	}

	// Seed stack items first (needed for relations)
	log.Println("Seeding stack items...")
	stackMap, err := s.seedStackItems(dryRun)
	if err != nil {
		return fmt.Errorf("failed to seed stack items: %w", err)
	}
	s.transformer.SetStackMap(stackMap)
	log.Printf("Created %d stack items", len(stackMap))

	// Seed profile
	log.Println("Seeding profile...")
	if err := s.seedProfile(dryRun); err != nil {
		return fmt.Errorf("failed to seed profile: %w", err)
	}

	// Seed experiences
	log.Println("Seeding experiences...")
	if err := s.seedExperiences(dryRun); err != nil {
		return fmt.Errorf("failed to seed experiences: %w", err)
	}

	// Seed education
	log.Println("Seeding education...")
	if err := s.seedEducation(dryRun); err != nil {
		return fmt.Errorf("failed to seed education: %w", err)
	}

	// Seed projects (combine projects and sideProjects)
	log.Println("Seeding projects...")
	allProjects := append(s.resume.Projects, s.resume.SideProjects...)
	if err := s.seedProjects(allProjects, dryRun); err != nil {
		return fmt.Errorf("failed to seed projects: %w", err)
	}

	log.Println("Seeding completed successfully!")
	return nil
}

// clearAllCollections deletes all records from all collections
func (s *Seeder) clearAllCollections() error {
	collections := []string{"profile", "experiences", "education", "projects", "stack"}

	for _, collectionName := range collections {
		// Get all records
		records, err := s.app.FindRecordsByFilter(collectionName, "", "", 1000, 0)
		if err != nil {
			log.Printf("Warning: failed to fetch records from %s: %v", collectionName, err)
			continue
		}

		// Delete each record
		for _, record := range records {
			if err := s.app.Delete(record); err != nil {
				log.Printf("Warning: failed to delete record %s from %s: %v", record.Id, collectionName, err)
			}
		}

		log.Printf("Cleared %d records from %s", len(records), collectionName)
	}

	return nil
}

// seedStackItems creates stack items from skills and returns a name-to-ID map
func (s *Seeder) seedStackItems(dryRun bool) (map[string]string, error) {
	stackRecords, err := s.transformer.TransformStackItems(s.resume.Skills)
	if err != nil {
		return nil, err
	}

	nameToID := make(map[string]string)

	if dryRun {
		log.Printf("Would create %d stack items", len(stackRecords))
		for _, record := range stackRecords {
			name := record.GetString("name")
			log.Printf("  - %s (category: %s)", name, record.GetString("category"))
		}
		return nameToID, nil
	}

	for _, record := range stackRecords {
		if err := s.app.Save(record); err != nil {
			return nil, fmt.Errorf("failed to save stack item %s: %w", record.GetString("name"), err)
		}
		nameToID[record.GetString("name")] = record.Id
	}

	return nameToID, nil
}

// seedProfile creates the profile record
func (s *Seeder) seedProfile(dryRun bool) error {
	record, err := s.transformer.TransformProfile(s.resume.Basics)
	if err != nil {
		return err
	}

	if dryRun {
		log.Printf("Would create profile: %s (%s)", record.GetString("full_name"), record.GetString("email"))
		return nil
	}

	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// seedExperiences creates experience records
func (s *Seeder) seedExperiences(dryRun bool) error {
	records, err := s.transformer.TransformExperiences(s.resume.Work)
	if err != nil {
		return err
	}

	if dryRun {
		log.Printf("Would create %d experience records", len(records))
		for _, record := range records {
			log.Printf("  - %s at %s", record.GetString("position"), record.GetString("company"))
		}
		return nil
	}

	for _, record := range records {
		if err := s.app.Save(record); err != nil {
			return fmt.Errorf("failed to save experience %s: %w", record.GetString("company"), err)
		}
	}

	return nil
}

// seedEducation creates education records
func (s *Seeder) seedEducation(dryRun bool) error {
	records, err := s.transformer.TransformEducation(s.resume.Education)
	if err != nil {
		return err
	}

	if dryRun {
		log.Printf("Would create %d education records", len(records))
		for _, record := range records {
			log.Printf("  - %s at %s", record.GetString("degree"), record.GetString("institution"))
		}
		return nil
	}

	for _, record := range records {
		if err := s.app.Save(record); err != nil {
			return fmt.Errorf("failed to save education %s: %w", record.GetString("institution"), err)
		}
	}

	return nil
}

// seedProjects creates project records
func (s *Seeder) seedProjects(projects []Project, dryRun bool) error {
	records, err := s.transformer.TransformProjects(projects)
	if err != nil {
		return err
	}

	if dryRun {
		log.Printf("Would create %d project records", len(records))
		for _, record := range records {
			log.Printf("  - %s (slug: %s)", record.GetString("name"), record.GetString("slug"))
		}
		return nil
	}

	for _, record := range records {
		if err := s.app.Save(record); err != nil {
			return fmt.Errorf("failed to save project %s: %w", record.GetString("name"), err)
		}
	}

	return nil
}
