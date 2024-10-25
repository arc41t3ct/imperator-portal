package handlers

import (
	"context"
	"imperatorapp/models"
	"net/http"

	"github.com/arc41t3ct/imperator"
)

type Handlers struct {
	App    *imperator.Imperator
	Models models.Models
}

// Convenience functions we can use in our handlers

// appName - return the app name
func (h *Handlers) appName() string {
	return h.App.AppName
}

// render - renders a page using the template engine defined in .env RENDERER
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, tmpl string, variables, data interface{}) error {
	return h.App.Render.Page(w, r, tmpl, variables, data)
}

// renderJet - renders a page using the jet template engine
func (h *Handlers) renderJet(w http.ResponseWriter, r *http.Request, tmpl string, variables, data interface{}) error {
	return h.App.Render.JetPage(w, r, tmpl, variables, data)
}

// renderGo - render a page using the go template engine
func (h *Handlers) renderGo(w http.ResponseWriter, r *http.Request, tmpl string, variables, data interface{}) error {
	return h.App.Render.GoPage(w, r, tmpl, variables, data)
}

func (h *Handlers) download(w http.ResponseWriter, r *http.Request, path string, file string) error {
	return h.App.Render.DownloadFile(w, r, path, file)
}

func (h *Handlers) renderJSON(w http.ResponseWriter, data interface{}, status int, headers ...http.Header) error {
	return h.App.Render.WriteJSON(w, data, status, headers...)
}

func (h *Handlers) renderXML(w http.ResponseWriter, data interface{}, status int, headers ...http.Header) error {
	return h.App.Render.WriteXML(w, data, status, headers...)
}

// sessionPut  - puts a new key value pair into the session given a context first
func (h *Handlers) sessionPut(ctx context.Context, key string, val interface{}) {
	h.App.Session.Put(ctx, key, val)
}

// sessionPut  - checks key to see if it is in the session given a context first
func (h *Handlers) sessionHas(ctx context.Context, key string) bool {
	return h.App.Session.Exists(ctx, key)
}

// sessionGet  - get a key our of the session given a context first
func (h *Handlers) sessionGet(ctx context.Context, key string) interface{} {
	return h.App.Session.Get(ctx, key)
}

// sessionRemove - removes a key out of the session fiven the context first
func (h *Handlers) sessionRemove(ctx context.Context, key string) {
	h.App.Session.Remove(ctx, key)
}

// sessionRenew - renews a session given the context
func (h *Handlers) sessionRenew(ctx context.Context) error {
	return h.App.Session.RenewToken(ctx)
}

// sessionDestroy - destroys a session
func (h *Handlers) sessionDestroy(ctx context.Context) error {
	return h.App.Session.Destroy(ctx)
}

// randomString - return a randomly generated string give its length with n
func (h *Handlers) randomString(n int) string {
	return h.App.CreateRadomString(n)
}

func (h *Handlers) encrypt(text string) (string, error) {
	enc := imperator.Encryption{Key: []byte(h.App.EncryptionKey)}
	encrypted, err := enc.Encrypt(text)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func (h *Handlers) decrypt(encryptedText string) (string, error) {
	enc := imperator.Encryption{Key: []byte(h.App.EncryptionKey)}
	decrypted, err := enc.Decrypt(encryptedText)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}
