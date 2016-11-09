package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/viper"
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
	rootPath      = "journal"
	splitSentence = regexp.MustCompile("^(.*?[.?!])\\s*(.*)$")
	validPath     = regexp.MustCompile(("^/" + rootPath + "/(?:save/)?(\\w{1,20})$"))
)

type Entry struct {
	Date  time.Time
	Title string
	Body  string
}

func (e Entry) String() string {
	return fmt.Sprintf("%s %s\n%s", e.Date.Format("2006-01-02 15:04"), e.Title, e.Body)
}

type Journal struct {
	Name    string
	Entries []Entry
}

func (j *Journal) Save() error {
	file, err := os.Create("journals/" + j.Name + ".txt")
	log.Println("---- Opening " + j.Name + ".txt ----")
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	for _, entry := range j.Entries {
		n, err := file.WriteString(entry.String())
		file.WriteString("\n\n")
		log.Printf("Wrote %d bytes. Err: %s\n", n, err)
	}
	file.Sync()
	return nil
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

		// Add the entry to the entries
		journal.Entries = append(journal.Entries, entry)
	}
	return &journal, nil
}

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
