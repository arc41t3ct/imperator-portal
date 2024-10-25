package middleware

import (
	"fmt"
	"imperatorapp/models"
	"net/http"
	"time"

	"github.com/arc41t3ct/imperator"
)

type Middleware struct {
	App    *imperator.Imperator
	Models models.Models
}

// deleteRememberCookie deletes the remeber cookie
func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remeber", m.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   m.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)
	m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
}
