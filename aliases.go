package main

import "net/http"

// Here we store aliases to imperator deep nested functions as a convenience

func (a *application) get(p string, hf http.HandlerFunc) {
	a.App.Routes.Get(p, hf)
}

func (a *application) post(p string, hf http.HandlerFunc) {
	a.App.Routes.Post(p, hf)
}

func (a *application) delete(p string, hf http.HandlerFunc) {
	a.App.Routes.Delete(p, hf)
}

func (a *application) use(m ...func(http.Handler) http.Handler) {
	a.App.Routes.Use(m...)
}
