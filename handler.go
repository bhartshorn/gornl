package main

import (
	"log"
	"net/http"
)

type journalHandler struct {
	journals *JournalDB
}

func (h journalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, name, err := parseUrl(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(regexUrl.String())
		log.Println(r.URL.Path)
	}

	journal, err := h.journals.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	switch method {
	case "view":
		viewHandler(w, r, &journal)
	}
	w.Write([]byte("You want to " + method + " the journal " + name))
	return
}
