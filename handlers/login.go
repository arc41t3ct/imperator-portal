package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"imperatorapp/models"
	"net/http"
	"time"

	jet "github.com/CloudyKit/jet/v6"
)

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	variables := make(jet.VarMap)
	variables.Set("error", "")
	if err := h.render(w, r, "login", variables, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.Session.Put(r.Context(), "error", fmt.Sprintf("failed to parse form with err: %s", err))
		http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		h.App.Session.Put(r.Context(), "error", "login failed")
		http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		h.App.Session.Put(r.Context(), "error", "login failed")
		http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
		return
	}

	if !matches {
		h.App.Session.Put(r.Context(), "error", "login failed")
		http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
		return
	}
	// did the user check the remember me?
	if r.Form.Get("remember") == "remember" {
		randomStr := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomStr))
		if err != nil {
			h.App.Render.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := models.RememberToken{
			UserID:        user.ID,
			RememberToken: sha,
		}
		_, err = h.Models.RememberToken.Insert(rm)
		if err != nil {
			h.App.Session.Put(r.Context(), "error", "login failed")
			http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
			return
		}

		// set cookie
		expire := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315350000, // 1 yr
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}
	h.App.Session.Put(r.Context(), "userID", user.ID)
	h.App.Session.Put(r.Context(), "success", fmt.Sprintf("Welcome %s %s in the admin area", user.FirstName, user.LastName))
	http.Redirect(w, r, "/admin/area", http.StatusSeeOther)
}
