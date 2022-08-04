// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/bindata"
)

type DataFunc func(*http.Request) interface{}

const mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

var usernameRegex = regexp.MustCompile(`[a-zA-z\-_0-9]+`)

// Views is a collection of templates
type Views struct {
	source            string
	e                 *Environment
	injectedData      map[string]interface{}
	injectedDataFuncs map[string]DataFunc

	templates map[string]*template.Template
}

// NewViews reads a set of templates from a directory and loads them
// into a Views. Custom functions are injected into the templates.
func NewViews(e *Environment, basepath string) (v *Views, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	v = &Views{
		source:            basepath,
		e:                 e,
		injectedData:      make(map[string]interface{}),
		injectedDataFuncs: make(map[string]DataFunc),
		templates:         make(map[string]*template.Template),
	}
	if err := v.loadTemplates(); err != nil {
		return nil, err
	}
	return v, nil
}

func customTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"list": func(values ...interface{}) ([]interface{}, error) {
			return values, nil
		},
		"plus1": func(a int) int {
			return a + 1
		},
		"sub1": func(a int) int {
			return a - 1
		},
		"titleBool": func(b bool) string {
			if b {
				return "True"
			}
			return "False"
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"isUsername": func(s string) bool {
			return usernameRegex.Match([]byte(s))
		},
	}
}

func (v *Views) loadTemplates() error {
	dir := v.source

	mainTemplate, err := template.New("main").Parse(mainTmpl)
	if err != nil {
		return err
	}
	mainTemplate.Funcs(customTemplateFuncs())

	partialFiles, err := bindata.AssetDir(dir + "/partials")
	if err != nil {
		return errors.New("layout templates not found")
	}

	pages, err := bindata.AssetDir(dir + "/pages")
	if err != nil {
		return errors.New("templates not found")
	}

	for _, file := range pages {
		if !strings.HasSuffix(file, ".tmpl") {
			continue
		}

		fileName := filepath.Base(file)
		fileName = fileName[:len(fileName)-5]

		v.e.Log.WithField("filename", fileName).Debug("Loading page template")

		pageTemplate, _ := mainTemplate.Clone()

		// Partial templates available to all pages
		for _, partial := range partialFiles {
			fileBytes, _ := bindata.GetAsset(dir + "/partials/" + partial)
			if _, err := pageTemplate.Parse(string(fileBytes)); err != nil {
				v.e.Log.WithFields(verbose.Fields{
					"partial": partial,
					"error":   err.Error(),
				}).Debug("Failed to load partial template")
				return err
			}
		}

		// Specific layout for this page
		pageLayout := strings.SplitN(fileName, "-", 2)[0]
		fileBytes, _ := bindata.GetAsset(dir + "/layouts/" + pageLayout + ".tmpl")
		if _, err := pageTemplate.Parse(string(fileBytes)); err != nil {
			v.e.Log.WithFields(verbose.Fields{
				"layout": pageLayout,
				"error":  err.Error(),
			}).Debug("Failed to load layout template")
			return err
		}

		// This page template
		fileBytes, _ = bindata.GetAsset(dir + "/pages/" + file)
		if _, err := pageTemplate.Parse(string(fileBytes)); err != nil {
			v.e.Log.WithFields(verbose.Fields{
				"file":  file,
				"error": err.Error(),
			}).Debug("Failed to load page template")
			return err
		}

		v.templates[fileName] = pageTemplate
	}
	return nil
}

// NewView returns a template associated with a request.
func (v *Views) NewView(file string, r *http.Request) *View {
	return &View{
		name:              file,
		t:                 v.templates[file],
		e:                 v.e,
		r:                 r,
		injectedData:      v.injectedData,
		injectedDataFuncs: v.injectedDataFuncs,
	}
}

// InjectData will always inject a specific key, value pair into every template
func (v *Views) InjectData(key string, val interface{}) {
	v.injectedData[key] = val
}

// InjectData will always inject a specific key, value pair into every template
func (v *Views) InjectDataFunc(key string, fn DataFunc) {
	v.injectedDataFuncs[key] = fn
}

// Reload replaces the Views object with a new one using the same source
// directory. DO NOT call this function in production. This functions should
// only be called in a development environment. This functions is very
// susceptible to race conditions.
func (v *Views) Reload() error {
	views, err := NewViews(v.e, v.source)
	if err != nil {
		return err
	}
	v.templates = views.templates
	return nil
}

// RenderError renders an error template with the given data. If data is nil,
// the generic "error" template is used. If data is not nil, "custom-error"
// is used.
func (v *Views) RenderError(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
	if data == nil {
		v.NewView("user-error", r).Render(w, nil)
		return
	}
	v.NewView("user-custom-error", r).Render(w, data)
}

// View represents a template associated with a specific request.
type View struct {
	name              string
	t                 *template.Template
	e                 *Environment
	r                 *http.Request
	injectedData      map[string]interface{}
	injectedDataFuncs map[string]DataFunc
}

type FlashMessageType int

const (
	FlashMessageSuccess FlashMessageType = iota
	FlashMessageWarning
	FlashMessageError
)

type FlashMessage struct {
	Message string
	Type    FlashMessageType
}

func (f FlashMessageType) String() string {
	switch f {
	case FlashMessageSuccess:
		return "success"
	case FlashMessageWarning:
		return "warning"
	case FlashMessageError:
		return "error"
	}
	return ""
}

// Render executes the template and writes it to w. This function will also
// save a web session to the client if "username" is set in the current session.
func (v *View) Render(w http.ResponseWriter, data map[string]interface{}) {
	if v.t == nil {
		v.e.Log.WithFields(verbose.Fields{
			"template": v.name,
			"package":  "common:views",
		}).Error("Error rendering template, v.t is nil")
		return
	}

	if data == nil {
		data = make(map[string]interface{})
	}
	session := GetSessionFromContext(v.r)
	flashes := session.Flashes()
	var flashMessage FlashMessage
	if len(flashes) > 0 {
		flashMessage = flashes[0].(FlashMessage)
	}

	if session.GetString("username") != "" {
		if err := session.Save(v.r, w); err != nil {
			v.e.Log.WithField("error", err).Error("Failed to save session")
		}
	}

	data["flashMessage"] = template.HTML(flashMessage.Message)
	data["flashMessageType"] = flashMessage.Type.String()
	data["layout"] = strings.SplitN(v.name, "-", 2)[0]

	for key, val := range v.injectedData {
		data[key] = val
	}
	for key, fn := range v.injectedDataFuncs {
		data[key] = fn(v.r)
	}

	if err := v.t.Execute(w, data); err != nil {
		v.e.Log.WithFields(verbose.Fields{
			"template": v.name,
			"error":    err,
			"package":  "common:views",
		}).Error("Error rendering template")
	}
}
