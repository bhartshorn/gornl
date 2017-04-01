package main

import (
	"errors"
	"github.com/spf13/viper"
	"gopkg.in/hlandau/passlib.v1"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	regexUrl = &regexp.Regexp{}
)

func init() {
	// This is an ugly hack. Not sure how to do it better as of now, but I
	// have to set this variable here to get the viper variable.
	regexUrl = regexp.MustCompile(("^/" + viper.GetString("ServerPrefix") + "/(save|view|test)/(\\w{1,20})$"))
}

type journalHandler struct {
	journals *JournalDB
}

func (h journalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, name, err := parseUrl(r)
	prefix := "/" + viper.GetString("ServerPrefix")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Bad URL: " + r.URL.Path)
		log.Println(regexUrl.String())
		return
	}

	journal, err := h.journals.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	user, pass, hasAuth := r.BasicAuth()
	if !hasAuth {
		log.Println("Requesting auth for journal " + name)
		w.Header().Set("WWW-Authenticate", "Basic realm="+name+"\"")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	newHash, authErr := passlib.Verify(pass, journal.Password)

	// Ewwww... code duplication. I'm sorry, Mr Torvalds... code smell
	if !hasAuth || user != journal.Username || authErr != nil {
		log.Println("Unauthorized access attempt to " + name)
		w.Header().Set("WWW-Authenticate", "Basic realm="+name+"\"")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if newHash != "" {
		h.journals.ChangePass(name, newHash)
		log.Println("Pass changed")
	}

	switch method {
	case "view":
		viewHandler(w, r, &journal)
	case "save":
		r.ParseForm()
		for k, v := range r.Form {
			log.Println(k + ": " + strings.Join(v, ""))
		}
		switch r.FormValue("submit") {
		case "Save Entry":
			err := h.journals.Add(name, r.FormValue("body"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, prefix+"/view/"+name, http.StatusFound)
		case "Change Password":
			if r.FormValue("password") != r.FormValue("confirm") {
				http.Error(w, "Passwords did not match!", http.StatusInternalServerError)
				return
			}
			encPass, err := passlib.Hash(r.FormValue("password"))
			if err != nil {
				http.Error(w, "Could not hash password", http.StatusInternalServerError)
			}
			h.journals.ChangePass(name, encPass)
			log.Println("New password: " + r.FormValue("password") + " : " + encPass)
		}

		http.Redirect(w, r, prefix+"/view/"+name, http.StatusFound)
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
