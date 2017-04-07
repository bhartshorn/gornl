package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	regexPass     = regexp.MustCompile("^password: (.*)$")
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
	Password string
}

func (j *Journal) Save() error {
	file, err := os.Create(viper.GetString("JournalPath") + "/" + j.Name + ".txt")
	defer file.Sync()
	log.Println("---- Opening " + j.Name + ".txt ----")
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	file.WriteString("password: " + j.Password)
	for _, entry := range j.Entries {
		_, err := file.WriteString(entry.String() + "\n\n")
		if err != nil {
			return err
		}
	}
	return nil
}

type JournalDB struct {
	mu       sync.Mutex
	journals map[string]*Journal
}

func (db *JournalDB) Get(name string) (Journal, error) {
	journal, err := db.get(name)
	return *journal, err
}

func (db *JournalDB) get(name string) (*Journal, error) {
	// If the journal is already in our "database", send it by value
	if journal, open := db.journals[name]; open {
		return journal, nil
	}

	// or we need to open it, then send it
	err := db.loadJournal(name)
	if err != nil {
		return &Journal{}, err
	}

	if journal, open := db.journals[name]; open {
		return journal, nil
	}

	return &Journal{}, nil
}

func (db *JournalDB) loadJournal(name string) error {
	file, err := os.Open(viper.GetString("JournalPath") + "/" + name + ".txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	journal := Journal{Name: name}

	scanner.Scan()
	journal.Password = regexPass.FindStringSubmatch(scanner.Text())[1]
	if len(journal.Password) == 0 {
		return fmt.Errorf("Journal password is empty")
	}

	// While Scanner can open a new line
	for scanner.Scan() {
		// At the beginning of this loop, should always be the beginning of an entry
		entry := Entry{}

		entry.Date, err = time.Parse("2006-01-02 15:04", scanner.Text()[0:16])
		if err != nil {
			return fmt.Errorf("Issue parsing date in entry")
		}
		entry.Title = scanner.Text()[17:len(scanner.Text())]
		if len(entry.Title) == 0 {
			return fmt.Errorf("Title of entry is empty")
		}

		// We got the date & title, now get the body. It may be on
		// multiple lines.
		for scanner.Scan() && len(scanner.Text()) != 0 {
			// If it is in multiple lines, we will probably need to add a space
			if len(entry.Body) > 0 {
				entry.Body += " "
			}
			entry.Body += scanner.Text()
		}

		if len(entry.Body) == 0 {
			return fmt.Errorf("Entry body is empty")
		}

		// Add the entry to the entries
		journal.Entries = append(journal.Entries, entry)
	}

	db.journals[name] = &journal
	return nil
}

func (db *JournalDB) Put(journal Journal) error {
	db.mu.Lock()
	db.journals[journal.Name] = &journal
	db.journals[journal.Name].Save()
	db.mu.Unlock()
	return nil
}

func (db *JournalDB) Add(name string, rawEntry string) error {
	journal, err := db.get(name)
	split := regexSentence.FindStringSubmatch(rawEntry)
	if len(split) != 3 {
		return fmt.Errorf("Couldn't parse entry text")
	}
	entry := Entry{time.Now(), split[1], split[2]}
	db.mu.Lock()
	journal.Entries = append(journal.Entries, entry)
	journal.Save()
	db.mu.Unlock()

	return err
}

func (db *JournalDB) ChangePass(name string, pass string) error {
	journal, err := db.get(name)
	journal.Password = pass
	db.mu.Lock()
	journal.Save()
	db.mu.Unlock()

	return err
}

func (db *JournalDB) load(name string) {
	return
}

func newJournalDB() JournalDB {
	return JournalDB{
		journals: make(map[string]*Journal),
	}
}
