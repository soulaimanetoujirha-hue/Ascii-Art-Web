package main

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"ascii-art-web/ascii"
)

const maxInputLength = 1000

// banners lists the allowed banner names, so users cannot ask for other files.
var banners = map[string]bool{
	"standard":   true,
	"shadow":     true,
	"thinkertoy": true,
}

// pageData is what the index template displays.
type pageData struct {
	Text   string
	Banner string
	Result string
	Error  string
}

// homeHandler serves the main page on GET /.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	render(w, http.StatusOK, pageData{Banner: "standard"})
}

// asciiArtHandler turns the posted text into ASCII art on POST /ascii-art.
func asciiArtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	data := pageData{
		Text:   r.FormValue("text"),
		Banner: r.FormValue("banner"),
	}

	// A mistake in the form: show the page again with a message
	if message := checkInput(data.Text, data.Banner); message != "" {
		data.Error = message
		render(w, http.StatusBadRequest, data)
		return
	}

	result, err := ascii.Render(data.Text, filepath.Join("banners", data.Banner+".txt"))
	switch {
	case errors.Is(err, ascii.ErrInvalidChar):
		data.Error = "Only printable ASCII characters are allowed."
		render(w, http.StatusBadRequest, data)
		return
	case errors.Is(err, ascii.ErrBannerNotFound):
		log.Println(err)
		http.Error(w, "404 Not Found: banner", http.StatusNotFound)
		return
	case err != nil:
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Result = result
	render(w, http.StatusOK, data)
}

// checkInput returns a message for the user when the form is wrong, or "" when it is fine.
func checkInput(text, banner string) string {
	switch {
	case text == "":
		return "Please type some text."
	case len(text) > maxInputLength:
		return "Text is too long."
	case !banners[banner]:
		return "Please choose a valid banner."
	}
	return ""
}

// render fills the index template with data and sends it with status.
func render(w http.ResponseWriter, status int, data pageData) {
	// The template is read on every request, so a missing file gives a 404
	tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		log.Println(err)
		http.Error(w, "404 Not Found: template", http.StatusNotFound)
		return
	}

	// Execute into a buffer first so a failure can still send a clean 500
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}
