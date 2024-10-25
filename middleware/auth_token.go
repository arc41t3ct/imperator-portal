package middleware

import "net/http"

func (m *Middleware) AuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := m.Models.Tokens.AuthenticationToken(r)
		if err != nil {
			var payload struct {
				Error   bool   `json:"error"`
				Message string `json:"message"`
			}
			payload.Error = true
			payload.Message = "unauthorized"

			if err := m.App.Render.WriteJSON(w, payload, http.StatusUnauthorized); err != nil {
				m.App.ErrorLog.Println("failed to write json with err:", err)
			}
		}
	})
}
