package handlers

import (
	"fmt"
	"imperatorapp/models"
	"net/http"

	jet "github.com/CloudyKit/jet/v6"
	"github.com/arc41t3ct/imperator/mailer"
	"github.com/arc41t3ct/imperator/signer"
)

// PasswordForgot handles request for people who forgot their passwords
func (h *Handlers) PasswordForgot(w http.ResponseWriter, r *http.Request) {
	h.App.InfoLog.Println("running handler: PasswordForgot")
	if err := h.render(w, r, "password_forgot", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.Render.Error404(w, r)
	}
}

// PasswordForgotPost handles request for people who forgot their passwords
func (h *Handlers) PasswordForgotPost(w http.ResponseWriter, r *http.Request) {
	h.App.InfoLog.Println("running handler: PasswordForgotPost")
	// parse form
	if err := r.ParseForm(); err != nil {
		h.App.Render.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// verify email exists
	var u *models.User
	email := r.Form.Get("email")
	u, err := u.GetByEmail(email)
	if err != nil {
		h.App.Render.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// create link to password reset form
	link := fmt.Sprintf("%s/admin/users/reset?email=%s", h.App.Server.URL, email)
	// sign the link with our signer
	sign := signer.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	signedLink := sign.GenerateTokenFromString(link)
	h.App.InfoLog.Println("signed new link:", signedLink)
	// email the message
	var data struct {
		Link     string
		FirsName string
	}
	data.Link = signedLink
	data.FirsName = u.FirstName
	msg := mailer.Message{
		To:       u.Email,
		Subject:  "Password Reset for " + email,
		Template: "password_reset",
		Data:     data,
		From:     "admin@imperator.portal",
	}
	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		h.App.Render.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// redirect the user
	h.App.Session.Put(
		r.Context(),
		"flash",
		"Check your inbox for a reset password link.")
	http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
}

// PasswordReset handles request for resetting a password
func (h *Handlers) PasswordReset(w http.ResponseWriter, r *http.Request) {
	h.App.InfoLog.Println("running handler: PasswordReset")
	// get form values
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testUrl := fmt.Sprintf("%s%s", h.App.Server.URL, theURL)
	// validate url and not tempered
	signer := signer.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	valid := signer.VerifyToken(testUrl)
	if !valid {
		h.App.ErrorLog.Println("invalid url")
		h.App.Render.ErrorUnauthorized(w, r)
	}
	// make sure it is not expired
	expired := signer.Expired(testUrl, 60)
	if expired {
		h.App.ErrorLog.Print("expired url")
		h.App.Render.ErrorUnauthorized(w, r)
	}
	// display form
	encryptedEmail, err := h.encrypt(email)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.Render.Error404(w, r)
	}
	vars := make(jet.VarMap)
	vars.Set("email", encryptedEmail)

	if err := h.render(w, r, "password_reset", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.Render.Error404(w, r)
	}
}

// PasswordResetPost handles request for resetting passesords
func (h *Handlers) PasswordResetPost(w http.ResponseWriter, r *http.Request) {
	h.App.InfoLog.Println("running handler: PasswordResetPost")
	// parse form
	if err := r.ParseForm(); err != nil {
		h.App.Render.Error500(w, r)
		return
	}
	// get and decrypt the email
	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.Session.Put(
			r.Context(),
			"error",
			"The could not reset the password.")
		http.Redirect(w, r, "/admin/users/reset-password", http.StatusSeeOther)
		return
	}
	// get the user
	var u models.User
	user, err := u.GetByEmail(email)
	if err != nil {
		h.App.Session.Put(
			r.Context(),
			"error",
			"The could not reset the password.")
		http.Redirect(w, r, "/admin/users/reset-password", http.StatusSeeOther)
		return
	}
	// reset Password
	if err := user.ResetPassword(user.ID, r.Form.Get("password")); err != nil {
		h.App.Render.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// redirect
	h.App.Session.Put(
		r.Context(),
		"flash",
		"The password was reset successfully. You can log in with the new one now.")
	http.Redirect(w, r, "/admin/users/login", http.StatusSeeOther)
}
