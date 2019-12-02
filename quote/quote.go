package quote

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/oddlid/bot"
	"io"
	"io/ioutil"
	"time"
)

const (
	POINTS_MAX   int    = 6
	QUOTE_FILE string = "/tmp/quotes.json"
)

type IRCMsg struct {
	TS      time.Time `json:"timestamp"`
	Channel string    `json:"ircchannel"`
	Nick    string    `json:"nick"`
	Msg     string    `json:"message"`
}

type MsgCache struct {
	BufSize int      `json:"bufsize"`
	Entries []IRCMsg `json:"entries"`
}

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

func NewMsgCache(bufsize int) *MsgCache {
	return &MsgCache{
		BufSize: bufsize,
		Entries: make([]IRCMsg, 0, bufsize),
	}
}

func (mc *MsgCache) Len() int {
	return len(mc.Entries)
}

//func (mc *MsgCache) NextFree() int {
//	l := mc.Len()
//	if l < mc.BufSize {
//		return l
//	}
//	return l
//}

func (mc *MsgCache) Full() bool {
	return mc.Len() == mc.BufSize
}

func (mc *MsgCache) Add(msg IRCMsg) {
	if mc.Full() {
		// shift entries
		for i := 0; i < mc.Len()-1; i++ {
			mc.Entries[i] = mc.Entries[i+1]
		}
	}
	mc.Entries[mc.Len()-1] = msg
}

// This is in serious need of getting looked at while sober...
func (mc *MsgCache) GetRange(rollback, howmany int) ([]IRCMsg, error) {
	msgs := IRCMsg{}
	idx := mc.Len() - rollback - 1
	if idx < 0 {
		return msgs, fmt.Errorf("Bad rollback index: %d", rollback)
	}
	if (idx + howmany) >= mc.Len() { // bad math, drunk...
		return msgs, fmt.Errorf("Fuck off, I should be sleeping...")
	}
	return mc.Entries[idx:howmany] // maybe not right at all
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
