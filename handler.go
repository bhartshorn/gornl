package main

import (
	"errors"
	"log"
	"net/http"
	"time"
)

var (
	regexUrl = regexp.MustCompile(("^/" + rootPath + "/(save|view|test)/(\\w{1,20})$"))
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
	}

	switch method {
	case "view":
		viewHandler(w, r, &journal)
		return
	case "save":
		split := regexSentence.FindStringSubmatch(r.FormValue("body"))
		entry := Entry{time.Now(), split[1], split[2]}
		journal.Entries = append(journal.Entries, entry)
		h.journals.Put(journal)
		http.Redirect(w, r, "/"+rootPath+"/view/"+name, http.StatusFound)
		return
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
