package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"

	"github.com/damione1/personal-website/internal/handlers"
	"github.com/damione1/personal-website/internal/seeds"
	"github.com/damione1/personal-website/internal/services"
	_ "github.com/damione1/personal-website/pb_migrations"
)

var (
	Version    string = "dev"
	CommitHash string = "unknown"
	BuildDate  string = "unknown"
)

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Personal Website v%s\n", Version)
		fmt.Printf("Commit: %s\n", CommitHash)
		fmt.Printf("Built: %s\n", BuildDate)
		os.Exit(0)
	}

	app := pocketbase.New()

	// Register migrate command with automigrate enabled
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true, // Auto-run migrations on app.Start()
	})

	// Register seed command
	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed database from jsoncv resume JSON file",
		Long:  "Load resume data from a jsoncv-formatted JSON file and populate PocketBase collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			resumePath, _ := cmd.Flags().GetString("resume")
			if resumePath == "" {
				return fmt.Errorf("--resume flag is required")
			}

			clearExisting, _ := cmd.Flags().GetBool("clear")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Initialize app (bootstrap) without starting server
			if err := app.Bootstrap(); err != nil {
				return fmt.Errorf("failed to bootstrap app: %w", err)
			}

			seeder := seeds.NewSeeder(app)
			if err := seeder.Seed(resumePath, clearExisting, dryRun); err != nil {
				return err
			}

			return nil
		},
	}
	seedCmd.Flags().String("resume", "", "Path to jsoncv resume JSON file (required)")
	seedCmd.Flags().Bool("clear", true, "Clear existing data before seeding")
	seedCmd.Flags().Bool("dry-run", false, "Validate and show what would be created without making changes")
	app.RootCmd.AddCommand(seedCmd)

	// Initialize services
	contentManager := services.NewContentManager(app)

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(contentManager)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Page routes
		se.Router.GET("/", homeHandler.Home)
		se.Router.GET("/about", homeHandler.About)

		// Experience routes
		se.Router.GET("/experience", homeHandler.ExperienceList)

		// Projects routes
		se.Router.GET("/projects", homeHandler.ProjectList)
		se.Router.GET("/projects/{slug}", homeHandler.ProjectDetail)

		// Blog routes
		se.Router.GET("/blog", homeHandler.BlogList)
		se.Router.GET("/blog/{slug}", homeHandler.BlogDetail)

		// Stack page
		se.Router.GET("/stack", homeHandler.StackPage)

		// Static files - must be registered last with wildcard path
		se.Router.GET("/static/{path...}", apis.Static(os.DirFS("./web/static"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
