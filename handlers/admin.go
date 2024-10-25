package handlers

import (
	"net/http"
)

// Admin handles request for
func (h *Handlers) Admin(w http.ResponseWriter, r *http.Request) {
	h.App.InfoLog.Println("running handler: Admin")
	if err := h.render(w, r, "admin", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}
