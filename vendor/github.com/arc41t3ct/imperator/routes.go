package imperator

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (i *Imperator) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.CleanPath)
	mux.Use(middleware.Recoverer)
	if i.Debug {
		mux.Use(middleware.Logger)
	}
	mux.Use(i.SessionLoad)
	mux.Use(i.NoSurf)

	return mux
}
