package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/damione1/personal-website/internal/handlers"
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
