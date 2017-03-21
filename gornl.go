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

func init() {
	viper.AddConfigPath("config")
	viper.SetConfigName("gornl")
	viper.SetDefault("ServerPort", "8080")
	viper.SetDefault("ServerPrefix", "journal")
	viper.SetDefault("JournalPath", "journals")
}

func main() {
	journals := newJournalDB()

	prefix := "/" + viper.GetString("ServerPrefix") + "/"
	port := ":" + viper.GetString("ServerPort")

	http.Handle(prefix, journalHandler{&journals})
	http.Handle(prefix+"static/",
		http.StripPrefix(prefix+"static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(port, nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, j *Journal) {
	err := templates.ExecuteTemplate(w, tmpl+".html", j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
