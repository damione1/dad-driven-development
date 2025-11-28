package main

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/apis"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	app := pocketbase.New()

	// Register routes
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Homepage
		se.Router.GET("/", func(e *core.RequestEvent) error {
			return e.String(200, "Personal Website - Coming Soon!")
		})

		// Static files
		se.Router.GET("/static/{path...}", apis.Static(os.DirFS("./web/static"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
