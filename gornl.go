package main

import (
	"errors"
	"github.com/spf13/viper"
	"html/template"
	"net/http"
	"regexp"
	"time"
)

var (
	templates = template.Must(template.ParseFiles(
		"tmpl/header.html",
		"tmpl/footer.html",
		"tmpl/edit.html",
		"tmpl/view.html"))
	rootPath      = "journal"
	splitSentence = regexp.MustCompile("^(.*?[.?!])\\s*(.*)$")
	validPath     = regexp.MustCompile(("^/" + rootPath + "/(?:save/)?(\\w{1,20})$"))
)

func main() {
	viper.AddConfigPath("config")
	viper.SetConfigName("gornl")
	viper.SetDefault("ServerPort", "8080")
	viper.SetDefault("ServerPath", "journal")

	http.HandleFunc("/"+rootPath+"/save/", saveHandler)
	http.HandleFunc("/"+rootPath+"/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, j *Journal) {
	err := templates.ExecuteTemplate(w, tmpl+".html", j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(w, r)
	if err != nil {
		return
	}
	j, err := loadJournal(name)
	if err != nil {
		http.Redirect(w, r, "/edit/"+name, http.StatusFound)
		return
	}
	renderTemplate(w, "view", j)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	_, err := getName(w, r)
	if err != nil {
		return
	}
	//p, err := loadPage(name)
	if err != nil {
		//p = &Page{Name: name}
	}
	//renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	journal, err := loadJournal(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	split := splitSentence.FindStringSubmatch(r.FormValue("body"))

	entry := Entry{time.Now(), split[1], split[2]}

	journal.Entries = append(journal.Entries, entry)

	journal.Save()

	http.Redirect(w, r, "/"+rootPath+"/"+name, http.StatusFound)
}

func getName(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Journal Name")
	}
	return m[1], nil
}
