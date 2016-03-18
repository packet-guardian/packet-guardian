package common

import (
	"fmt"
	"net/http"
)

// NotImplementedHandler is a mock handler for paths that aren't implemented yet
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("The path \"%s\" is not implemented yet\n", r.URL.Path)
	w.Write([]byte(resp))
}

// StringInSlice searches a slice for a string
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
