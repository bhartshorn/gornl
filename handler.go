package main

import (
	"errors"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"regexp"
)

var (
	regexUrl = regexp.MustCompile(("^/" + viper.GetString("ServerPath") + "/(save|view|test)/(\\w{1,20})$"))
)

type journalHandler struct {
	journals *JournalDB
}

func (h journalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, name, err := parseUrl(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Bad URL: " + r.URL.Path)
		return
	}

	journal, err := h.journals.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println("Can't find journal \"" + name + "\"")
		return
	}

	user, pass, hasAuth := r.BasicAuth()

	if !hasAuth || user != journal.Username || pass != journal.Password {
		w.Header().Set("WWW-Authenticate", "Basic realm="+name)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		log.Println("Unauthorized access attempt to " + name)
		return
	}

	switch method {
	case "view":
		viewHandler(w, r, &journal)
	case "save":
		err := h.journals.Add(name, r.FormValue("body"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/"+viper.GetString("ServerPath")+"/view/"+name, http.StatusFound)
	}
	log.Println("You want to " + method + " the journal " + name)
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request, j *Journal) {
	renderTemplate(w, "view", j)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
}

func parseUrl(r *http.Request) (string, string, error) {
	matches := regexUrl.FindStringSubmatch(r.URL.Path)
	if len(matches) != 3 {
		return "", "", errors.New("Invalid URL")
	}
	method := matches[1]
	name := matches[2]

	return method, name, nil
}
