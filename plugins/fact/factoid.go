// © 2013 the CatBase Authors under the WTFPL. See AUTHORS for the list of authors.

package fact

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/velour/catbase/bot"
	"github.com/velour/catbase/bot/msg"
)

// The factoid plugin provides a learning system to the bot so that it can
// respond to queries in a way that is unpredictable and fun

// factoid stores info about our factoid for lookup and later interaction
type factoid struct {
	id       sql.NullInt64
	Fact     string
	Tidbit   string
	Verb     string
	Owner    string
	created  time.Time
	accessed time.Time
	Count    int
}

type alias struct {
	Fact string
	Next string
}

func (a *alias) resolve(db *sqlx.DB) (*factoid, error) {
	// perform DB query to fill the To field
	q := `select fact, next from factoid_alias where fact=?`
	var next alias
	err := db.Get(&next, q, a.Next)
	if err != nil {
		// we hit the end of the chain, get a factoid named Next
		fact, err := getSingleFact(db, a.Next)
		if err != nil {
			err := fmt.Errorf("Error resolvig alias %v: %v", a, err)
			return nil, err
		}
		return fact, nil
	}
	return next.resolve(db)
}

func findAlias(db *sqlx.DB, fact string) (bool, *factoid) {
	q := `select * from factoid_alias where fact=?`
	var a alias
	err := db.Get(&a, q, fact)
	if err != nil {
		return false, nil
	}
	f, err := a.resolve(db)
	return err == nil, f
}

func (a *alias) save(db *sqlx.DB) error {
	q := `select * from factoid_alias where fact=?`
	var offender alias
	err := db.Get(&offender, q, a.Next)
	if err == nil {
		return fmt.Errorf("DANGER: an opposite alias already exists")
	}
	_, err = a.resolve(db)
	if err != nil {
		return fmt.Errorf("there is no fact at that destination")
	}
	q = `insert or replace into factoid_alias (fact, next) values (?, ?)`
	_, err = db.Exec(q, a.Fact, a.Next)
	if err != nil {
		return err
	}
	return nil
}

func aliasFromStrings(from, to string) *alias {
	return &alias{from, to}
}

func (f *factoid) save(db *sqlx.DB) error {
	var err error
	if f.id.Valid {
		// update
		_, err = db.Exec(`update factoid set
			fact=?,
			tidbit=?,
			verb=?,
			owner=?,
			accessed=?,
			count=?
		where id=?`,
			f.Fact,
			f.Tidbit,
			f.Verb,
			f.Owner,
			f.accessed.Unix(),
			f.Count,
			f.id.Int64)
	} else {
		f.created = time.Now()
		f.accessed = time.Now()
		// insert
		res, err := db.Exec(`insert into factoid (
			fact,
			tidbit,
			verb,
			owner,
			created,
			accessed,
			count
		) values (?, ?, ?, ?, ?, ?, ?);`,
			f.Fact,
			f.Tidbit,
			f.Verb,
			f.Owner,
			f.created.Unix(),
			f.accessed.Unix(),
			f.Count,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		// hackhackhack?
		f.id.Int64 = id
		f.id.Valid = true
	}
	return err
}

func (f *factoid) delete(db *sqlx.DB) error {
	var err error
	if f.id.Valid {
		_, err = db.Exec(`delete from factoid where id=?`, f.id)
	}
	f.id.Valid = false
	return err
}

func getFacts(db *sqlx.DB, fact string, tidbit string) ([]*factoid, error) {
	var fs []*factoid
	query := `select
			id,
			fact,
			tidbit,
			verb,
			owner,
			created,
			accessed,
			count
		from factoid
		where fact like ?
		and tidbit like ?;`
	rows, err := db.Query(query,
		"%"+fact+"%", "%"+tidbit+"%")
	if err != nil {
		log.Printf("Error regexping for facts: %s", err)
		return nil, err
	}
	for rows.Next() {
		var f factoid
		var tmpCreated int64
		var tmpAccessed int64
		err := rows.Scan(
			&f.id,
			&f.Fact,
			&f.Tidbit,
			&f.Verb,
			&f.Owner,
			&tmpCreated,
			&tmpAccessed,
			&f.Count,
		)
		if err != nil {
			return nil, err
		}
		f.created = time.Unix(tmpCreated, 0)
		f.accessed = time.Unix(tmpAccessed, 0)
		fs = append(fs, &f)
	}
	return fs, err
}

func getSingle(db *sqlx.DB) (*factoid, error) {
	var f factoid
	var tmpCreated int64
	var tmpAccessed int64
	err := db.QueryRow(`select
			id,
			fact,
			tidbit,
			verb,
			owner,
			created,
			accessed,
			count
		from factoid
		order by random() limit 1;`).Scan(
		&f.id,
		&f.Fact,
		&f.Tidbit,
		&f.Verb,
		&f.Owner,
		&tmpCreated,
		&tmpAccessed,
		&f.Count,
	)
	f.created = time.Unix(tmpCreated, 0)
	f.accessed = time.Unix(tmpAccessed, 0)
	return &f, err
}

func getSingleFact(db *sqlx.DB, fact string) (*factoid, error) {
	var f factoid
	var tmpCreated int64
	var tmpAccessed int64
	err := db.QueryRow(`select
			id,
			fact,
			tidbit,
			verb,
			owner,
			created,
			accessed,
			count
		from factoid
		where fact like ?
		order by random() limit 1;`,
		fact).Scan(
		&f.id,
		&f.Fact,
		&f.Tidbit,
		&f.Verb,
		&f.Owner,
		&tmpCreated,
		&tmpAccessed,
		&f.Count,
	)
	f.created = time.Unix(tmpCreated, 0)
	f.accessed = time.Unix(tmpAccessed, 0)
	return &f, err
}

// Factoid provides the necessary plugin-wide needs
type Factoid struct {
	Bot      bot.Bot
	NotFound []string
	LastFact *factoid
	db       *sqlx.DB
}

// NewFactoid creates a new Factoid with the Plugin interface
func New(botInst bot.Bot) *Factoid {
	p := &Factoid{
		Bot: botInst,
		NotFound: []string{
			"I don't know.",
			"NONONONO",
			"((",
			"*pukes*",
			"NOPE! NOPE! NOPE!",
			"One time, I learned how to jump rope.",
		},
		db: botInst.DB(),
	}

	if _, err := p.db.Exec(`create table if not exists factoid (
			id integer primary key,
			fact string,
			tidbit string,
			verb string,
			owner string,
			created integer,
			accessed integer,
			count integer
		);`); err != nil {
		log.Fatal(err)
	}

	if _, err := p.db.Exec(`create table if not exists factoid_alias (
			fact string,
			next string,
			primary key (fact, next)
		);`); err != nil {
		log.Fatal(err)
	}

	for _, channel := range botInst.Config().Channels {
		go p.factTimer(channel)

		go func(ch string) {
			// Some random time to start up
			time.Sleep(time.Duration(15) * time.Second)
			if ok, fact := p.findTrigger(p.Bot.Config().Factoid.StartupFact); ok {
				p.sayFact(msg.Message{
					Channel: ch,
					Body:    "speed test", // BUG: This is defined in the config too
					Command: true,
					Action:  false,
				}, *fact)
			}
		}(channel)
	}

	return p
}

// findAction simply regexes a string for the action verb
func findAction(message string) string {
	r, err := regexp.Compile("<.+?>")
	if err != nil {
		panic(err)
	}
	action := r.FindString(message)

	if action == "" {
		if strings.Contains(message, " is ") {
			return "is"
		} else if strings.Contains(message, " are ") {
			return "are"
		}
	}

	return action
}

// learnFact assumes we have a learning situation and inserts a new fact
// into the database
func (p *Factoid) learnFact(message msg.Message, fact, verb, tidbit string) bool {
	verb = strings.ToLower(verb)

	var count sql.NullInt64
	err := p.db.QueryRow(`select count(*) from factoid
		where fact=? and verb=? and tidbit=?`,
		fact, verb, tidbit).Scan(&count)
	if err != nil {
		log.Println("Error counting facts: ", err)
		return false
	} else if count.Valid && count.Int64 != 0 {
		log.Println("User tried to relearn a fact.")
		return false
	}

	n := factoid{
		Fact:     fact,
		Tidbit:   tidbit,
		Verb:     verb,
		Owner:    message.User.Name,
		created:  time.Now(),
		accessed: time.Now(),
		Count:    0,
	}
	p.LastFact = &n
	err = n.save(p.db)
	if err != nil {
		log.Println("Error inserting fact: ", err)
		return false
	}

	return true
}

// findTrigger checks to see if a given string is a trigger or not
func (p *Factoid) findTrigger(fact string) (bool, *factoid) {
	fact = strings.ToLower(fact) // TODO: make sure this needs to be lowered here

	f, err := getSingleFact(p.db, fact)
	if err != nil {
		return findAlias(p.db, fact)
	}
	return true, f
}

// sayFact spits out a fact to the channel and updates the fact in the database
// with new time and count information
func (p *Factoid) sayFact(message msg.Message, fact factoid) {
	msg := p.Bot.Filter(message, fact.Tidbit)
	full := p.Bot.Filter(message, fmt.Sprintf("%s %s %s",
		fact.Fact, fact.Verb, fact.Tidbit,
	))
	for i, m := 0, strings.Split(msg, "$and"); i < len(m) && i < 4; i++ {
		msg := strings.TrimSpace(m[i])
		if len(msg) == 0 {
			continue
		}

		if fact.Verb == "action" {
			p.Bot.SendAction(message.Channel, msg)
		} else if fact.Verb == "reply" {
			p.Bot.SendMessage(message.Channel, msg)
		} else {
			p.Bot.SendMessage(message.Channel, full)
		}
	}

	// update fact tracking
	fact.accessed = time.Now()
	fact.Count += 1
	err := fact.save(p.db)
	if err != nil {
		log.Printf("Could not update fact.\n")
		log.Printf("%#v\n", fact)
		log.Println(err)
	}
	p.LastFact = &fact
}

// trigger checks the message for its fitness to be a factoid and then hauls
// the message off to sayFact for processing if it is in fact a trigger
func (p *Factoid) trigger(message msg.Message) bool {
	minLen := p.Bot.Config().Factoid.MinLen
	if len(message.Body) > minLen || message.Command || message.Body == "..." {
		if ok, fact := p.findTrigger(message.Body); ok {
			p.sayFact(message, *fact)
			return true
		}
		r := strings.NewReplacer("'", "", "\"", "", ",", "", ".", "", ":", "",
			"?", "", "!", "")
		if ok, fact := p.findTrigger(r.Replace(message.Body)); ok {
			p.sayFact(message, *fact)
			return true
		}
	}

	return false
}

// tellThemWhatThatWas is a hilarious name for a function.
func (p *Factoid) tellThemWhatThatWas(message msg.Message) bool {
	fact := p.LastFact
	var msg string
	if fact == nil {
		msg = "Nope."
	} else {
		msg = fmt.Sprintf("That was (#%d) '%s <%s> %s'",
			fact.id.Int64, fact.Fact, fact.Verb, fact.Tidbit)
	}
	p.Bot.SendMessage(message.Channel, msg)
	return true
}

func (p *Factoid) learnAction(message msg.Message, action string) bool {
	body := message.Body

	parts := strings.SplitN(body, action, 2)
	// This could fail if is were the last word or it weren't in the sentence (like no spaces)
	if len(parts) != 2 {
		return false
	}

	trigger := strings.TrimSpace(parts[0])
	fact := strings.TrimSpace(parts[1])
	action = strings.TrimSpace(action)

	if len(trigger) == 0 || len(fact) == 0 || len(action) == 0 {
		p.Bot.SendMessage(message.Channel, "I don't want to learn that.")
		return true
	}

	if len(strings.Split(fact, "$and")) > 4 {
		p.Bot.SendMessage(message.Channel, "You can't use more than 4 $and operators.")
		return true
	}

	strippedaction := strings.Replace(strings.Replace(action, "<", "", 1), ">", "", 1)

	if p.learnFact(message, trigger, strippedaction, fact) {
		p.Bot.SendMessage(message.Channel, fmt.Sprintf("Okay, %s.", message.User.Name))
	} else {
		p.Bot.SendMessage(message.Channel, "I already know that.")
	}

	return true
}

// Checks body for the ~= operator returns it
func changeOperator(body string) string {
	if strings.Contains(body, "=~") {
		return "=~"
	} else if strings.Contains(body, "~=") {
		return "~="
	}
	return ""
}

// If the user requesting forget is either the owner of the last learned fact or
// an admin, it may be deleted
func (p *Factoid) forgetLastFact(message msg.Message) bool {
	if p.LastFact == nil {
		p.Bot.SendMessage(message.Channel, "I refuse.")
		return true
	}

	err := p.LastFact.delete(p.db)
	if err != nil {
		log.Println("Error removing fact: ", p.LastFact, err)
	}
	fmt.Printf("Forgot #%d: %s %s %s\n", p.LastFact.id.Int64, p.LastFact.Fact,
		p.LastFact.Verb, p.LastFact.Tidbit)
	p.Bot.SendAction(message.Channel, "hits himself over the head with a skillet")
	p.LastFact = nil

	return true
}

// Allow users to change facts with a simple regexp
func (p *Factoid) changeFact(message msg.Message) bool {
	oper := changeOperator(message.Body)
	parts := strings.SplitN(message.Body, oper, 2)
	userexp := strings.TrimSpace(parts[1])
	trigger := strings.TrimSpace(parts[0])

	parts = strings.Split(userexp, "/")

	log.Printf("changeFact: %s %s %#v", trigger, userexp, parts)

	if len(parts) == 4 {
		// replacement
		if parts[0] != "s" {
			p.Bot.SendMessage(message.Channel, "Nah.")
		}
		find := parts[1]
		replace := parts[2]

		// replacement
		result, err := getFacts(p.db, trigger, parts[1])
		if err != nil {
			log.Println("Error getting facts: ", trigger, err)
		}
		if userexp[len(userexp)-1] != 'g' {
			result = result[:1]
		}
		// make the changes
		msg := fmt.Sprintf("Changing %d facts.", len(result))
		p.Bot.SendMessage(message.Channel, msg)
		reg, err := regexp.Compile(find)
		if err != nil {
			p.Bot.SendMessage(message.Channel, "I don't really want to.")
			return false
		}
		for _, fact := range result {
			fact.Fact = reg.ReplaceAllString(fact.Fact, replace)
			fact.Fact = strings.ToLower(fact.Fact)
			fact.Verb = reg.ReplaceAllString(fact.Verb, replace)
			fact.Tidbit = reg.ReplaceAllString(fact.Tidbit, replace)
			fact.Count += 1
			fact.accessed = time.Now()
			fact.save(p.db)
		}
	} else if len(parts) == 3 {
		// search for a factoid and print it
		result, err := getFacts(p.db, trigger, parts[1])
		if err != nil {
			log.Println("Error getting facts: ", trigger, err)
			p.Bot.SendMessage(message.Channel, "bzzzt")
			return true
		}
		count := len(result)
		if count == 0 {
			p.Bot.SendMessage(message.Channel, "I didn't find any facts like that.")
			return true
		}
		if parts[2] == "g" && len(result) > 4 {
			// summarize
			result = result[:4]
		} else {
			p.sayFact(message, *result[0])
			return true
		}
		msg := fmt.Sprintf("%s ", trigger)
		for i, fact := range result {
			if i != 0 {
				msg = fmt.Sprintf("%s |", msg)
			}
			msg = fmt.Sprintf("%s <%s> %s", msg, fact.Verb, fact.Tidbit)
		}
		if count > 4 {
			msg = fmt.Sprintf("%s | ...and %d others", msg, count)
		}
		p.Bot.SendMessage(message.Channel, msg)
	} else {
		p.Bot.SendMessage(message.Channel, "I don't know what you mean.")
	}
	return true
}

// Message responds to the bot hook on recieving messages.
// This function returns true if the plugin responds in a meaningful way to the users message.
// Otherwise, the function returns false and the bot continues execution of other plugins.
func (p *Factoid) Message(message msg.Message) bool {
	if strings.ToLower(message.Body) == "what was that?" {
		return p.tellThemWhatThatWas(message)
	}

	// This plugin has no business with normal messages
	if !message.Command {
		// look for any triggers in the db matching this message
		return p.trigger(message)
	}

	if strings.HasPrefix(strings.ToLower(message.Body), "alias") {
		log.Printf("Trying to learn an alias: %s", message.Body)
		m := strings.TrimPrefix(message.Body, "alias ")
		parts := strings.SplitN(m, "->", 2)
		if len(parts) != 2 {
			p.Bot.SendMessage(message.Channel, "If you want to alias something, use: `alias this -> that`")
			return true
		}
		a := aliasFromStrings(strings.TrimSpace(parts[1]), strings.TrimSpace(parts[0]))
		if err := a.save(p.db); err != nil {
			p.Bot.SendMessage(message.Channel, err.Error())
		} else {
			p.Bot.SendAction(message.Channel, "learns a new synonym")
		}
		return true
	}

	if strings.ToLower(message.Body) == "factoid" {
		if fact := p.randomFact(); fact != nil {
			p.sayFact(message, *fact)
			return true
		}
		log.Println("Got a nil fact.")
	}

	if strings.ToLower(message.Body) == "forget that" {
		return p.forgetLastFact(message)
	}

	if changeOperator(message.Body) != "" {
		return p.changeFact(message)
	}

	action := findAction(message.Body)
	if action != "" {
		return p.learnAction(message, action)
	}

	// look for any triggers in the db matching this message
	if p.trigger(message) {
		return true
	}

	// We didn't find anything, panic!
	p.Bot.SendMessage(message.Channel, p.NotFound[rand.Intn(len(p.NotFound))])
	return true
}

// Help responds to help requests. Every plugin must implement a help function.
func (p *Factoid) Help(channel string, parts []string) {
	p.Bot.SendMessage(channel, "I can learn facts and spit them back out. You can say \"this is that\" or \"he <has> $5\". Later, trigger the factoid by just saying the trigger word, \"this\" or \"he\" in these examples.")
	p.Bot.SendMessage(channel, "I can also figure out some variables including: $nonzero, $digit, $nick, and $someone.")
}

// Empty event handler because this plugin does not do anything on event recv
func (p *Factoid) Event(kind string, message msg.Message) bool {
	return false
}

// Pull a fact at random from the database
func (p *Factoid) randomFact() *factoid {
	f, err := getSingle(p.db)
	if err != nil {
		fmt.Println("Error getting a fact: ", err)
		return nil
	}
	return f
}

// factTimer spits out a fact at a given interval and with given probability
func (p *Factoid) factTimer(channel string) {
	duration := time.Duration(p.Bot.Config().Factoid.QuoteTime) * time.Minute
	myLastMsg := time.Now()
	for {
		time.Sleep(time.Duration(5) * time.Second) // why 5?

		lastmsg, err := p.Bot.LastMessage(channel)
		if err != nil {
			// Probably no previous message to time off of
			continue
		}

		tdelta := time.Since(lastmsg.Time)
		earlier := time.Since(myLastMsg) > tdelta
		chance := rand.Float64()
		success := chance < p.Bot.Config().Factoid.QuoteChance

		if success && tdelta > duration && earlier {
			fact := p.randomFact()
			if fact == nil {
				log.Println("Didn't find a random fact to say")
				continue
			}

			users := p.Bot.Who(channel)

			// we need to fabricate a message so that bot.Filter can operate
			message := msg.Message{
				User:    &users[rand.Intn(len(users))],
				Channel: channel,
			}
			p.sayFact(message, *fact)
			myLastMsg = time.Now()
		}
	}
}

// Handler for bot's own messages
func (p *Factoid) BotMessage(message msg.Message) bool {
	return false
}

// Register any web URLs desired
func (p *Factoid) RegisterWeb() *string {
	http.HandleFunc("/factoid/req", p.serveQuery)
	http.HandleFunc("/factoid", p.serveQuery)
	tmp := "/factoid"
	return &tmp
}

func linkify(text string) template.HTML {
	parts := strings.Fields(text)
	for i, word := range parts {
		if strings.HasPrefix(word, "http") {
			parts[i] = fmt.Sprintf("<a href=\"%s\">%s</a>", word, word)
		}
	}
	return template.HTML(strings.Join(parts, " "))
}

func (p *Factoid) serveQuery(w http.ResponseWriter, r *http.Request) {
	context := make(map[string]interface{})
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"linkify": linkify,
	}
	if e := r.FormValue("entry"); e != "" {
		entries, err := getFacts(p.db, e, "")
		if err != nil {
			log.Println("Web error searching: ", err)
		}
		context["Count"] = fmt.Sprintf("%d", len(entries))
		context["Entries"] = entries
		context["Search"] = e
	}
	t, err := template.New("factoidIndex").Funcs(funcMap).Parse(factoidIndex)
	if err != nil {
		log.Println(err)
	}
	err = t.Execute(w, context)
	if err != nil {
		log.Println(err)
	}
}

func (p *Factoid) ReplyMessage(message msg.Message, identifier string) bool { return false }
