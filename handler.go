package main

import (
	"log"
	"net/http"
)

type journalHandler struct {
	Journals *[]Journal
}

func (h journalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, name, err := parseUrl(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("test")
	}
	w.Write([]byte("You want to " + method + " the journal " + name))
	return
}
