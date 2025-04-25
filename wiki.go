package main

// https://go.dev/doc/articles/wiki/

import (
	"html/template"
	"os"
	"log"
	"net/http"
	"regexp"
	"path/filepath"
)

type Page struct {
	Title string
	Body []byte
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/edit.html", "templates/view.html"))

var validPath  = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	filename := p.Title + ".txt"

	return os.WriteFile("articles/" + filename, p.Body, 0600)
}

func load(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile("articles/" + filename)

	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)

	if err != nil {
		http.Error(w, "Could not load template: " + err.Error(), http.StatusInternalServerError)

		return
	}
}

func frontHandler(w http.ResponseWriter, r *http.Request) {
	entries, _ := os.ReadDir("articles")

	var articles []string

	for _, entry := range entries {
		title := entry.Name()[:len(entry.Name())-len(filepath.Ext(entry.Name()))]
		articles = append(articles, title)
	}

	err := templates.ExecuteTemplate(w, "index.html", articles)

	if err != nil {
		http.Error(w, "Could not the load template: " + err.Error(), http.StatusInternalServerError)

		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := load(title)

	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}
	
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := load(title)

	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()

	if err != nil {
		http.Error(w, "Error saving " + err.Error(), http.StatusFound)
	}

	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			
			return
		}

		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", frontHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":80", nil))
}