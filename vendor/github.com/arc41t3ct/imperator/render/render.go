package render

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	jet "github.com/CloudyKit/jet/v6"
	scs "github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type Render struct {
	Renderer         string
	RootPath         string
	Secure           bool
	Port             string
	ServerDomainName string
	JetViews         *jet.Set
	Session          *scs.SessionManager
}

type TemplateData struct {
	IsAuthenticated  bool
	Data             map[string]interface{}
	CSRFToken        string
	Port             string
	ServerDomainName string
	Secure           bool
	Error            string // stores error messages
	Flash            string // shows up for a short time
	Success          string // shows up for a short time
}

func (i *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = i.Secure
	td.ServerDomainName = i.ServerDomainName
	td.CSRFToken = nosurf.Token(r)
	td.Port = i.Port
	if i.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = true
	}
	td.Flash = i.Session.PopString(r.Context(), "flash")
	td.Error = i.Session.PopString(r.Context(), "error")
	td.Success = i.Session.PopString(r.Context(), "success")
	return td
}

func (i *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(i.Renderer) {
	case "go":
		return i.GoPage(w, r, view, variables, data)
	case "jet":
		return i.JetPage(w, r, view, variables, data)
	default:

	}
	return errors.New("Missing renderer")
}

// GoPage renders a standard go template
func (i *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", i.RootPath, view))
	if err != nil {
		return err
	}

	tmplData := &TemplateData{}
	if data != nil {
		tmplData = data.(*TemplateData)
	}

	err = tmpl.Execute(w, &tmplData)
	if err != nil {
		return err
	}

	return nil
}

// JetPage renders a template using the jet templating engine
func (i *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	var vars jet.VarMap
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = i.defaultData(td, r)

	t, err := i.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (i *Render) WriteJSON(w http.ResponseWriter, data interface{}, status int, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (i *Render) WriteXML(w http.ResponseWriter, data interface{}, status int, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	// out = []byte(fmt.Sprintf("<?xml-stylesheet href=\"%s\"?>\n%s", style, out))
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (i *Render) DownloadFile(w http.ResponseWriter, r *http.Request, pathTofile, fileName string) error {
	fp := path.Join(pathTofile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (i *Render) Error404(w http.ResponseWriter, r *http.Request) {
	i.ErrorStatus(w, http.StatusNotFound)
}

func (i *Render) Error500(w http.ResponseWriter, r *http.Request) {
	i.ErrorStatus(w, http.StatusInternalServerError)
}

func (i *Render) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	i.ErrorStatus(w, http.StatusUnauthorized)
}

func (i *Render) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
