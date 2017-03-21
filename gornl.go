package main

import (
	"github.com/spf13/viper"
	"html/template"
	"net/http"
)

var (
	templates = template.Must(template.ParseFiles(
		"tmpl/header.html",
		"tmpl/footer.html",
		"tmpl/edit.html",
		"tmpl/view.html"))
	rootPath = "journal"
)

func main() {
	viper.AddConfigPath("config")
	viper.SetConfigName("gornl")
	viper.SetDefault("ServerPort", "8080")
	viper.SetDefault("ServerPath", "journal")
	viper.SetDefault("JournalPath", "journals")

	journals := newJournalDB()

	http.Handle("/"+viper.GetString("ServerPath")+"/", journalHandler{&journals})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":"+viper.GetString("ServerPort"), nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, j *Journal) {
	err := templates.ExecuteTemplate(w, tmpl+".html", j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
