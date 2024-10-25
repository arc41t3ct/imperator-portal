package handlers

import (
	"fmt"
	"imperatorapp/models"
	"net/http"
	"time"
)

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// delete remeber token if exists
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := models.RememberToken{}
		if err := rt.DeleteByToken(h.App.Session.GetString(r.Context(), "remember_token")); err != nil {
			h.App.ErrorLog.Println("failed to delete remember token with err:", err)
		}
	}
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1, // 1 yr
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)
	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	h.App.Session.Remove(r.Context(), "remember_token")
	h.App.Session.Destroy(r.Context())
	h.App.Session.Put(r.Context(), "success", "You have been successfully logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
