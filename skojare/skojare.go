package skojare

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
	"io"
	"io/ioutil"
	"time"
)

const (
	POINTS_MAX   int    = 6
	SKOJBOT_FILE string = "/tmp/skojbot.json"
)

type Quote struct {
	Index   int       `json:"index"`
	Votes   int       `json:"votes"`
	Added   time.Time `json:"date_added"`
	AddedBy string    `json:"added_by"`
	Nick    string    `json:"nick"`
	Content string    `json:"content"`
}

type Quotes map[int]*Quote

type DB struct {
	Started time.Time `json:"db_started"`
	Entries Quotes    `json:"db_entries"`
}

func NewDB() *DB {
	return &DB{
		Started: time.Now(),
		Entries: make(Quotes),
	}
}

func NewQuote(bynick, nick, content string, idx int) *Quote {
	return &Quote{
		Index:   idx,
		Added:   time.Now(),
		AddedBy: bynick,
		Nick:    nick,
		Content: content,
	}
}

func (q *Quote) Vote(int points) int {
	// Only allow +-POINTS_MAX
	if points < -POINTS_MAX {
		points = -POINTS_MAX
	} else if points > POINTS_MAX {
		points = POINTS_MAX
	}
	q.Votes += points
	return q.Votes
}

func (q *Quote) String() string {
	return fmt.Sprintf("[#%d %v %s]: %s", q.Index, q.Added, q.Nick, q.Content)
	// #%d by %s, added by %s @ %v:
}

func (qs Quotes) Len() int {
	return len(qs)
}

func (qs Quotes) NextIndex() int {
	return qs.Len()
}

func (qs Quotes) Get(idx int) *Quote {
	item, found := qs[idx]
	if !found {
		return nil
	} else {
		return item
	}
}

func (qs Quotes) Add(bynick, nick, content string) {
	idx := qs.NextIndex()
	qs[idx] = NewQuote(bynick, nick, content, idx)
}

func (db *DB) ToJSON(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(db, "", "\t")
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (db *DB) FromJSON(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, db)
}

func list() (string, error) {
	return "", nil
}

func add() (string, error) {
	return "", nil
}

func vote() (string, error) {
	return "", nil
}

func skojare(cmd *bot.Cmd) (string, error) {
	log.Debugf("cmd.Args: %q", cmd.Args)
}

func init() {
	bot.RegisterCommand(
		"skojare",
		"Register a funny quote and let people vote for it",
		"list|add <nick> <content>|vote <index> <points>",
		skojare,
	)
}
