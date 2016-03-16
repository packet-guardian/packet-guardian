package common

import (
	"html/template"
)

var Templates *template.Template

func init() {
	Templates, _ = template.ParseGlob("templates/*.tmpl")
}
