package userwatch

/*
This plugin watches for irc users coming or going (join/part/quit).
Runs commands / sends messages when this happens.

NOTE: This plugin does not work as the normal plugins for "go-chat-bot",
      as this one needs a handle to both the bot instance and the ircevent.Connection
      instance, so just importing this prefixed with underscore and rely on init()
      will not work. We need to have a custom setup function, where we add callbacks
      to the ircevent.Connection instance. This should then be called by the importing
      package before irc.Run().

- Odd E. Ebbesen, 2019-02-07 18:32

*/

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/go-chat-bot/bot"
	"github.com/go-chat-bot/bot/irc"
	ircevent "github.com/thoj/go-ircevent"
)

const (
	JOIN        string = "JOIN"
	PART        string = "PART"
	QUIT        string = "QUIT"
	ADD         string = "ADD"
	DEL         string = "DEL"
	LS          string = "LS"
	RELOAD      string = "RELOAD"
	CLEAR       string = "CLEAR"
	PLUGIN      string = "UserWatch"
	DEF_CFGFILE string = "/tmp/userwatch.json"
)

var (
	_bot     *bot.Bot
	_cfg     *irc.Config
	_conn    *ircevent.Connection
	_wd      *WatchData
	_cfgfile string
)

type User struct {
	sync.RWMutex
	Nick string `json:"-"` // used for internal access, not needed in storage
	JMsg string `json:"jmsg,omitempty"`
	QMsg string `json:"qmsg,omitempty"`
}

type Channel struct {
	sync.RWMutex
	Users map[string]*User `json:"users"`
}

type WatchData struct {
	Modified time.Time           `json:"modified"`
	Channels map[string]*Channel `json:"channels"`
}

func InitBot(cfg *irc.Config, b *bot.Bot, conn *ircevent.Connection, cfgfile string) error {
	log.Debug("Initializing UserWatch...")
	_cfg = cfg
	_bot = b
	_conn = conn
	_cfgfile = cfgfile
	reload() // initializes _wd

	_conn.AddCallback(JOIN, onJOIN)
	_conn.AddCallback(QUIT, onQUIT)
	_conn.AddCallback(PART, onQUIT)

	register()

	return nil
}

func register() {
	// register command for interacting with this module
	// Arguments:
	//	add <nick> <join|part|quit> <msg>
	//	del <nick> [join|part|quit]
	//	ls  [nick] [join|part|quit]
	//	reload
	//	clear (not in doc output by design)
	m := []string{
		"nick",
		"JOIN|PART|QUIT",
	}
	argex := fmt.Sprintf(
		`arguments...
Where arguments can be one of:
  %s <%s> <%s> <msg>
  %s <%s> [%s]
  %s  [%s] [%s]
  %s

Examples:
  !userwatch %s MyNick %s Welcome, handsome %s
  !userwatch %s MyNick
  !userwatch %s MyNick
  !userwatch %s
`,
		ADD, m[0], m[1], DEL, m[0], m[1], LS, m[0], m[1], RELOAD, ADD, JOIN, "%s", DEL, LS, LS,
	)
	bot.RegisterCommand(
		"userwatch",
		"Display messages when users joins or quits/parts",
		argex,
		userwatch,
	)
}

func NewWatchData() *WatchData {
	return &WatchData{
		Modified: time.Now(),
		Channels: make(map[string]*Channel),
	}
}

func (wd *WatchData) Get(channel string) *Channel {
	c, found := wd.Channels[channel]
	if !found {
		log.Debugf("Creating channel %q with empty users", channel)
		c = &Channel{
			Users: make(map[string]*User),
		}
		wd.Channels[channel] = c
	}
	return c
}

func (wd *WatchData) Save(w io.Writer) (int, error) {
	wd.Modified = time.Now() // update timestamp
	jb, err := json.MarshalIndent(wd, "", "\t")
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (wd *WatchData) SaveFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := wd.Save(file)
	if err != nil {
		return err
	}
	log.Debugf("Saved %d bytes to %q", n, filename)
	return nil
}

func (wd *WatchData) Load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, wd)
}

func (wd *WatchData) LoadFile(filename string) *WatchData {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("WatchData.LoadFile(): Error opening %q - %s", filename, err.Error())
		return wd
	}
	defer file.Close()
	err = wd.Load(file)
	if err != nil {
		log.Error(err)
		return NewWatchData()
	}
	return wd
}

func (c *Channel) Has(nick string) bool {
	c.RLock()
	_, found := c.Users[nick]
	c.RUnlock()
	return found
}

func (c *Channel) Get(nick string) *User {
	c.RLock()
	u, found := c.Users[nick]
	c.RUnlock()
	if !found {
		log.Debugf("Creating new empty user %q", nick)
		u = &User{Nick: nick}
		c.Lock()
		c.Users[nick] = u
		c.Unlock()
	}
	return u
}

func (c *Channel) Nicks() []string {
	nicks := make([]string, 0, len(c.Users))
	for k := range c.Users {
		nicks = append(nicks, k)
	}
	sort.Strings(nicks)
	return nicks
}

func GetMsg(msg, nick string) string {
	if strings.Index(msg, "%s") > -1 && nick != "" {
		return fmt.Sprintf(msg, nick)
	}
	return msg
}

func (u *User) GetJMsg() string {
	return GetMsg(u.JMsg, u.Nick)
}

func (u *User) GetQMsg() string {
	return GetMsg(u.QMsg, u.Nick)
}

func (u *User) SetJMsg(msg string) {
	u.Lock()
	u.JMsg = msg
	u.Unlock()
}

func (u *User) SetQMsg(msg string) {
	u.Lock()
	u.QMsg = msg
	u.Unlock()
}

func onJOIN(e *ircevent.Event) {
	if e.Nick == _conn.GetNick() {
		log.Debugf("Seems it's myself joining. e.Nick: %s", e.Nick)
		return
	}

	c := _wd.Get(e.Arguments[0])
	if len(c.Users) == 0 {
		return
	}

	if !c.Has(e.Nick) {
		return
	}

	msg := c.Get(e.Nick).GetJMsg()
	if msg == "" {
		return
	}

	_bot.SendMessage(
		bot.OutgoingMessage{
			e.Arguments[0], // will be the channel name
			msg,
			&bot.User{
				ID:       e.Host,
				Nick:     e.Nick,
				RealName: e.User,
			},
			nil,
		},
	)
}

func onQUIT(e *ircevent.Event) {
	if e.Nick == _conn.GetNick() {
		log.Debugf("Seems it's myself leaving. e.Nick: %s", e.Nick)
		return
	}
	//log.Debugf("%s is leaving", e.Nick)

	c := _wd.Get(_cfg.Channels[0])
	if len(c.Users) == 0 {
		return
	}

	if !c.Has(e.Nick) {
		return
	}

	msg := c.Get(e.Nick).GetQMsg()
	if msg == "" {
		return
	}

	// maybe we should loop through all channels and send a msg for each here
	_bot.SendMessage(
		bot.OutgoingMessage{
			_cfg.Channels[0], // on QUIT e.Arguments[0] is empty, so we use this instead, even if not pretty
			msg,
			&bot.User{
				ID:       e.Host,
				Nick:     e.Nick,
				RealName: e.User,
			},
			nil,
		},
	)
}

func mtype(in, compare string) bool {
	return strings.ToUpper(in) == compare
}

func ls(channel, nick, msgtype string) string {
	c := _wd.Get(channel)
	if len(c.Users) == 0 {
		return fmt.Sprintf("%s: No configured messages for channel %q", PLUGIN, channel)
	}
	if nick == "" {
		return fmt.Sprintf("%s nicks: %s", PLUGIN, strings.Join(c.Nicks(), ", "))
	}
	if !c.Has(nick) {
		return fmt.Sprintf("%s: No configured messages for nick %q", PLUGIN, nick)
	}
	u := c.Get(nick)
	str := fmt.Sprintf("%s: messages for %s:\n", PLUGIN, nick)
	if mtype(msgtype, JOIN) {
		str += fmt.Sprintf("  %s: %s\n", JOIN, u.JMsg)
	} else if mtype(msgtype, QUIT) || mtype(msgtype, PART) {
		str += fmt.Sprintf("  %s: %s\n", QUIT, u.QMsg)
	} else {
		str += fmt.Sprintf("  %s: %s\n  %s: %s\n", JOIN, u.JMsg, QUIT, u.QMsg)
	}
	return str
}

func reload() {
	_wd = NewWatchData().LoadFile(_cfgfile) // will return new instance on error
}

func clear() {
	_wd = NewWatchData()
	_wd.SaveFile(_cfgfile)
}

func add(channel, nick, msgtype, msg string) (string, error) {
	if nick == "" || msgtype == "" || msg == "" {
		log.Error("add(): Empty nick, msgtype or msg")
		emsg := ADD + " Error: nick, message type and message has to be set"
		return emsg, fmt.Errorf(emsg)
	}

	c := _wd.Get(channel)
	u := c.Get(nick)

	ret := fmt.Sprintf("%s: Set/updated %s message for nick %q", PLUGIN, "%s", nick)
	if mtype(msgtype, JOIN) {
		u.SetJMsg(msg)
		ret = fmt.Sprintf(ret, JOIN)
	} else if mtype(msgtype, QUIT) || mtype(msgtype, PART) {
		u.SetQMsg(msg)
		ret = fmt.Sprintf(ret, QUIT)
	}
	_wd.SaveFile(_cfgfile)
	return ret, nil
}

func del(channel, nick, msgtype string) (string, error) {
	if nick == "" {
		log.Error("del(): Empty nick")
		emsg := "Error: empty nick"
		return emsg, fmt.Errorf(emsg)
	}

	c := _wd.Get(channel)
	u := c.Get(nick)

	cleanup := func() {
		if u.JMsg == "" && u.QMsg == "" {
			delete(c.Users, nick)
		}
	}

	ret := fmt.Sprintf("%s: Deleted %s message for nick %q", PLUGIN, "%s", nick)
	if mtype(msgtype, JOIN) {
		u.SetJMsg("")
		ret = fmt.Sprintf(ret, JOIN)
		cleanup()
	} else if mtype(msgtype, QUIT) || mtype(msgtype, PART) {
		u.SetQMsg("")
		ret = fmt.Sprintf(ret, QUIT)
		cleanup()
	} else if msgtype == "" {
		delete(c.Users, nick)
		ret = fmt.Sprintf("%s: Deleted nick %q", PLUGIN, nick)
	}

	_wd.SaveFile(_cfgfile)
	return ret, nil
}

func safeArgs(num int, args []string) []string {
	//log.Debugf("safeArgs: Got: %+v", args)
	alen := len(args)
	res := make([]string, num)
	for i := 0; i < num; i++ {
		//log.Debugf("safeArgs: Loop index: %d", i)
		if i < alen {
			res[i] = args[i]
		} else {
			res[i] = ""
		}
	}
	return res
}

// Handle runtime commands here
func userwatch(cmd *bot.Cmd) (string, error) {
	// Arguments:
	//	add <nick> <join|part|quit> <msg>
	//	del <nick> [join|part|quit]
	//	ls  [nick] [join|part|quit]
	//	reload
	//	clear ("secret")
	//
	// quit and part are synonymous.
	// del <nick> with no more args deletes the nick alltogether from the map.
	// ls <nick> with no more args shows both messages for join/quit.
	// ls with no more args shows a list of nicks that have messages set.
	// clear deletes everything without confirmation

	alen := len(cmd.Args)
	if alen == 0 {
		return "", nil
	}

	args := safeArgs(3, cmd.Args) // 3 is the longest possible set of args
	var retmsg string

	if mtype(args[0], LS) {
		return ls(cmd.Channel, args[1], args[2]), nil
	} else if mtype(args[0], ADD) {
		msg := strings.Join(cmd.Args[3:len(cmd.Args)], " ")
		return add(cmd.Channel, args[1], args[2], msg)
	} else if mtype(args[0], DEL) {
		return del(cmd.Channel, args[1], args[2])
	} else if mtype(args[0], CLEAR) {
		clear()
		retmsg = fmt.Sprintf("%s: DB cleared", PLUGIN)
	} else if mtype(args[0], RELOAD) {
		reload()
		retmsg = fmt.Sprintf("%s: DB reloaded from disk", PLUGIN)
	}

	return retmsg, nil
}
