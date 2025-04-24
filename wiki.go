package main

// https://go.dev/doc/articles/wiki/

import (
	"html/template"
	"os"
	"log"
	"net/http"
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"

	return os.WriteFile(filename, p.Body, 0600)
}

func load(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: body}, nil;
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles(tmpl + ".html")

	if err != nil {
		http.Error(w, "Error parsing template: " + err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, p);

	if err != nil {
		http.Error(w, "Error executing template: " + err.Error(), http.StatusInternalServerError)

		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := load(title)

	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}
	
	renderTemplate(w, "view", p);
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := load(title)

	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p);
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	//http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
