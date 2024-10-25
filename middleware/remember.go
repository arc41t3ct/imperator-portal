package middleware

import (
	"fmt"
	"imperatorapp/models"
	"net/http"
	"strconv"
	"strings"
)

// Remember Middleware allows a user to check remember me to be logged in after leaving
func (m *Middleware) Remember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			// user is not logged in
			cookie, err := r.Cookie(fmt.Sprintf("_%s_remeber", m.App.AppName))
			if err != nil {
				// no cookie, so on to the next middleware
				next.ServeHTTP(w, r)
			} else {
				// we found a cookie, so check it
				key := cookie.Value
				var u models.User
				if len(key) > 0 {
					split := strings.Split(key, "|")
					uId, hash := split[0], split[1]
					id, _ := strconv.Atoi(uId)
					validHash := u.CheckForRememberToken(id, hash)
					if !validHash {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "You have been logged out of the sessio")
						next.ServeHTTP(w, r)
					} else {
						// valid hash so log in user
						user, _ := u.Get(id)
						m.App.Session.Put(r.Context(), "userID", user.ID)
						m.App.Session.Put(r.Context(), "remember_token", hash)
					}
				} else {
					// key length is zero, so it's probably left over cookie (user no close browser)
					m.deleteRememberCookie(w, r)
					next.ServeHTTP(w, r)
				}
			}
		} else {
			// user is logged in
			next.ServeHTTP(w, r)
		}
	})
}
