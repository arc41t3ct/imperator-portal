package main

import (
	"imperatorapp/handlers"
	"imperatorapp/middleware"
	"imperatorapp/models"
	"log"
	"os"

	"github.com/arc41t3ct/imperator"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// init imperator
	imp := &imperator.Imperator{}
	err = imp.New(path)
	if err != nil {
		log.Fatal(err)
	}
	middle := &middleware.Middleware{}
	middle.App = imp
	hadls := &handlers.Handlers{}
	hadls.App = imp
	app := &application{}
	app.App = imp
	app.Middlware = middle
	app.Handlers = hadls
	app.App.Routes = app.routes()
	app.Models = models.New(app.App.DB.Pool)
	hadls.Models = app.Models
	middle.Models = app.Models

	return app
}
