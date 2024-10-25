package imperator

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (i *Imperator) SessionLoad(next http.Handler) http.Handler {
	return i.Session.LoadAndSave(next)
}

func (i *Imperator) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(i.config.cookie.secure)

	// exempt routes from csrf token protection
	csrfHandler.ExemptGlob("/api/cache/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   i.config.cookie.domain,
	})

	return csrfHandler
}
