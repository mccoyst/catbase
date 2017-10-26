package slack

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/velour/catbase/config"
)

type word struct {
	ID   int
	Word string
}

var s *Slack

func StartFix() {
	var cfile = flag.String("config", "config.lua",
		"Config file to load. (Defaults to config.lua)")
	flag.Parse() // parses the logging flags.

	c := config.Readconfig("1", *cfile)
	s = New(c)
	go s.Serve()
	select {
	case <-time.After(5 * time.Second):
		GetAllBroken(c.DBConn)
	}
	log.Println("done")
}

func GetAllBroken(db *sqlx.DB) {
	var words []word
	db.Select(&words, "select * from babblerWords where word like '<@%'")
	for _, w := range words {
		FixWord(db, w)
	}
}

func FixWord(db *sqlx.DB, w word) {
	// fix the text
	noAt := strings.TrimPrefix(w.Word, "<@")
	name := ""
	rest := ""
	ignore := false
	for i, c := range noAt {
		if c == '|' {
			ignore = true
		} else if c != '>' && !ignore {
			name = fmt.Sprintf("%s%c", name, c)
		} else if c == '>' {
			ignore = false
			rest = noAt[i+1 : len(noAt)]
			break
		}
	}

	// query slack
	nick, ok := s.getUser(strings.ToUpper(name))
	if !ok {
		log.Fatalf("Error querying for %s, %s", name, nick)
	}

	log.Printf("%s -> %s", w.Word, nick+rest)
	// save the id back

	_, err := db.Exec(`update babblerWords set word=? where id=?`, nick+rest, w.ID)
	if err != nil {
		log.Printf("Update error: %v", err)
	}
}
