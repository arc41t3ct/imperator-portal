package main

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes using aliases
	a.use(a.Middlware.Admin)
	a.use(a.Middlware.Remember)
	// routes go here using the aloases
	a.get("/", a.Handlers.Home)

	a.get("/admin/area", a.Handlers.Admin)
	a.get("/admin/user/login", a.Handlers.Login)
	a.post("/admin/user/login", a.Handlers.LoginPost)
	a.get("/admin/user/logout", a.Handlers.Logout)
	a.get("/admin/user/reset", a.Handlers.PasswordReset)
	a.get("/admin/user/forgot-password", a.Handlers.PasswordForgot)
	a.post("/admin/user/forgot-password", a.Handlers.PasswordForgotPost)
	a.get("/admin/user/reset-password", a.Handlers.PasswordReset)
	a.post("/admin/user/reset-password", a.Handlers.PasswordResetPost)

	// static routes do not edit below here
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))
	return a.App.Routes
}
