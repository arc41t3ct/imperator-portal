package main

import (
	"imperatorapp/handlers"
	"imperatorapp/middleware"
	"imperatorapp/models"

	"github.com/arc41t3ct/imperator"
)

type application struct {
	App       *imperator.Imperator
	Middlware *middleware.Middleware
	Handlers  *handlers.Handlers
	Models    models.Models
}

func main() {
	imp := initApplication()
	imp.App.ListenAndServe()
}
