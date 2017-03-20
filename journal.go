package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	regexUser     = regexp.MustCompile("^username: (.{1,32})$")
	regexPass     = regexp.MustCompile("^password: (.{1,32})$")
	regexSentence = regexp.MustCompile("^(.*?[.?!])\\s*(.*)$")
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
	Name     string
	Entries  []Entry
	Username string
	Password string
}

func (j *Journal) Save() error {
	file, err := os.Create("journals/" + j.Name + ".txt")
	log.Println("---- Opening " + j.Name + ".txt ----")
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("username: %s\n", j.Username))
	file.WriteString(fmt.Sprintf("password: %s\n", j.Password))
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

	scanner.Scan()
	journal.Username = regexUser.FindStringSubmatch(scanner.Text())[1]

	scanner.Scan()
	journal.Password = regexPass.FindStringSubmatch(scanner.Text())[1]

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

type JournalDB struct {
	mu       sync.Mutex
	journals map[string]*Journal
}

func (db *JournalDB) Get(name string) (Journal, error) {
	// If the journal is already in our "database", send it by value
	if journal, open := db.journals[name]; open {
		return *journal, nil
	}
	// or we need to open it, then send it
	journal, err := loadJournal(name)
	if err != nil {
		return Journal{}, err
	}

	db.journals[name] = journal

	return *journal, nil
}

func (db *JournalDB) Put(journal Journal) error {
	db.mu.Lock()
	db.journals[journal.Name] = &journal
	db.journals[journal.Name].Save()
	db.mu.Unlock()
	return nil
}

func (db *JournalDB) load(name string) {
	return
}

func newJournalDB() JournalDB {
	return JournalDB{
		journals: make(map[string]*Journal),
	}
}
