package main

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	templates = template.Must(template.ParseFiles(
		"tmpl/header.html",
		"tmpl/footer.html",
		"tmpl/edit.html",
		"tmpl/view.html"))
	rootPath  = "journal"
	validPath = regexp.MustCompile(("^/" + rootPath + "/(\\w{1,20})$"))
)

type Entry struct {
	Date  time.Time
	Title string
	Body  string
}

func (e Entry) String() string {
	return fmt.Sprintf("%s %s\n%s", e.Date, e.Title, e.Body)
}

type Journal struct {
	Name    string
	Entries []Entry
}

func loadJournal(name string) (*Journal, error) {
	file, err := os.Open("journals/" + name + ".txt")
	if err != nil {
		return &Journal{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	journal := Journal{Name: name}

	// While Scanner can open a new line
	for scanner.Scan() {
		// At the beginning of this loop, should always be the beginning of an entry
		entry := Entry{}

		entry.Date, err = time.Parse("2006-01-02 15:04", scanner.Text()[0:16])
		entry.Title = scanner.Text()[17:len(scanner.Text())]

		// We got the date & title, now get the body. It may be on
		// multiple lines.
		for scanner.Scan() && len(scanner.Text()) != 0 {
			// If it is in multiple lines, we will probably need to add a space
			if len(entry.Body) > 0 {
				entry.Body += " "
			}
			entry.Body += scanner.Text()
		}
		log.Println("----- New Entry ----")
		log.Println(entry)

		// Add the entry to the entries
		journal.Entries = append(journal.Entries, entry)
	}
	return &journal, nil
}
func main() {
	http.HandleFunc("/"+rootPath+"/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/go/", goHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, j *Journal) {
	err := templates.ExecuteTemplate(w, tmpl+".html", j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name != "" {
		http.Redirect(w, r, "/view/"+name, http.StatusFound)
	} else {
		http.Redirect(w, r, "/view/front", http.StatusFound)
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
		return
	}

	if r.FormValue("submit") == "Cancel" {
		http.Redirect(w, r, "/view/"+name, http.StatusFound)
	} else {
		body := r.FormValue("body")
		log.Println(body)
		http.Redirect(w, r, "/"+rootPath+"/"+name, http.StatusFound)
	}
}

func getName(w http.ResponseWriter, r *http.Request) (string, error) {
	log.Println(r.URL.Path)
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Journal Name")
	}
	return m[1], nil
}

func (j *Journal) save() error {
	return nil
}
