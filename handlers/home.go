package handlers

import (
	"net/http"
	"time"

	jet "github.com/CloudyKit/jet/v6"
)

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	defer h.App.LoadTime(time.Now())
	vars := make(jet.VarMap)
	vars.Set("app_name", h.appName())
	if err := h.render(w, r, "home", vars, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
		return
	}
}
