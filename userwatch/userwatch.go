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
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/go-chat-bot/bot/irc"
	"github.com/rs/zerolog/log"
	ircevent "github.com/thoj/go-ircevent"
)

const (
	cmdJoin   string = "JOIN"
	cmdPart   string = "PART"
	cmdQuit   string = "QUIT"
	cmdAdd    string = "ADD"
	cmdDel    string = "DEL"
	cmdList   string = "LS"
	cmdReload string = "RELOAD"
	cmdClear  string = "CLEAR"
	plugin    string = "UserWatch"
	// defaultConfigFile string = "/tmp/userwatch.json"
)

var (
	_bot     *bot.Bot
	_cfg     *irc.Config
	_conn    *ircevent.Connection
	_wd      *WatchData
	_cfgfile string
	_log     = log.With().Str("plugin", plugin).Logger()
)

type User struct {
	Nick string `json:"-"` // used for internal access, not needed in storage
	JMsg string `json:"jmsg,omitempty"`
	QMsg string `json:"qmsg,omitempty"`
	sync.RWMutex
}

type Channel struct {
	Users map[string]*User `json:"users"`
	sync.RWMutex
}

type WatchData struct {
	Modified time.Time           `json:"modified"`
	Channels map[string]*Channel `json:"channels"`
}

func InitBot(cfg *irc.Config, b *bot.Bot, conn *ircevent.Connection, cfgfile string) error {
	_log.Debug().
		Msg("Initializing UserWatch")
	_cfg = cfg
	_bot = b
	_conn = conn
	_cfgfile = cfgfile
	reload() // initializes _wd

	_conn.AddCallback(cmdJoin, onJOIN)
	_conn.AddCallback(cmdQuit, onQUIT)
	_conn.AddCallback(cmdPart, onQUIT)

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
		cmdAdd, m[0], m[1], cmdDel, m[0], m[1], cmdList, m[0], m[1], cmdReload, cmdAdd, cmdJoin, "%s", cmdDel, cmdList, cmdList,
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
		_log.Debug().
			Str("channel", channel).
			Msg("Creating channel with empty users")
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
	_log.Debug().
		Str("filename", filename).
		Int("bytes", n).
		Msg("Userwatch data saved")
	return nil
}

func (wd *WatchData) Load(r io.Reader) error {
	jb, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, wd)
}

func (wd *WatchData) LoadFile(filename string) *WatchData {
	file, err := os.Open(filename)
	if err != nil {
		_log.Error().
			Err(err).
			Str("filename", filename).
			Send()
		return wd
	}
	defer file.Close()
	err = wd.Load(file)
	if err != nil {
		_log.Error().
			Err(err).
			Send()
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
		_log.Debug().
			Str("nick", nick).
			Msg("Creating new, empty user")
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
	if strings.Contains(msg, "%s") && nick != "" {
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
		_log.Debug().
			Str("nick", e.Nick).
			Msg("Seems it's myself joining")
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
			Target:  e.Arguments[0], // will be the channel name
			Message: msg,
			Sender: &bot.User{
				ID:       e.Host,
				Nick:     e.Nick,
				RealName: e.User,
			},
		},
	)
}

func onQUIT(e *ircevent.Event) {
	if e.Nick == _conn.GetNick() {
		_log.Debug().
			Str("nick", e.Nick).
			Msg("Seems it's myself leaving")
		return
	}

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
			Target:  _cfg.Channels[0], // on QUIT e.Arguments[0] is empty, so we use this instead, even if not pretty
			Message: msg,
			Sender: &bot.User{
				ID:       e.Host,
				Nick:     e.Nick,
				RealName: e.User,
			},
		},
	)
}

func mtype(in, compare string) bool {
	return strings.ToUpper(in) == compare
}

func ls(channel, nick, msgtype string) string {
	c := _wd.Get(channel)
	if len(c.Users) == 0 {
		return fmt.Sprintf("%s: No configured messages for channel %q", plugin, channel)
	}
	if nick == "" {
		return fmt.Sprintf("%s nicks: %s", plugin, strings.Join(c.Nicks(), ", "))
	}
	if !c.Has(nick) {
		return fmt.Sprintf("%s: No configured messages for nick %q", plugin, nick)
	}
	u := c.Get(nick)
	str := fmt.Sprintf("%s: messages for %s:\n", plugin, nick)
	if mtype(msgtype, cmdJoin) {
		str += fmt.Sprintf("  %s: %s\n", cmdJoin, u.JMsg)
	} else if mtype(msgtype, cmdQuit) || mtype(msgtype, cmdPart) {
		str += fmt.Sprintf("  %s: %s\n", cmdQuit, u.QMsg)
	} else {
		str += fmt.Sprintf("  %s: %s\n  %s: %s\n", cmdJoin, u.JMsg, cmdQuit, u.QMsg)
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
		_log.Error().
			Str("func", "add()").
			Msg("Empty nick, msgtype or msg")
		emsg := cmdAdd + " Error: nick, message type and message has to be set"
		return emsg, fmt.Errorf(emsg)
	}

	c := _wd.Get(channel)
	u := c.Get(nick)

	ret := fmt.Sprintf("%s: Set/updated %s message for nick %q", plugin, "%s", nick)
	if mtype(msgtype, cmdJoin) {
		u.SetJMsg(msg)
		ret = fmt.Sprintf(ret, cmdJoin)
	} else if mtype(msgtype, cmdQuit) || mtype(msgtype, cmdPart) {
		u.SetQMsg(msg)
		ret = fmt.Sprintf(ret, cmdQuit)
	}
	_wd.SaveFile(_cfgfile)
	return ret, nil
}

func del(channel, nick, msgtype string) (string, error) {
	if nick == "" {
		emsg := "empty nick"
		_log.Error().
			Str("func", "del()").
			Msg(emsg)
		return emsg, fmt.Errorf(emsg)
	}

	c := _wd.Get(channel)
	u := c.Get(nick)

	cleanup := func() {
		if u.JMsg == "" && u.QMsg == "" {
			delete(c.Users, nick)
		}
	}

	ret := fmt.Sprintf("%s: Deleted %s message for nick %q", plugin, "%s", nick)
	if mtype(msgtype, cmdJoin) {
		u.SetJMsg("")
		ret = fmt.Sprintf(ret, cmdJoin)
		cleanup()
	} else if mtype(msgtype, cmdQuit) || mtype(msgtype, cmdPart) {
		u.SetQMsg("")
		ret = fmt.Sprintf(ret, cmdQuit)
		cleanup()
	} else if msgtype == "" {
		delete(c.Users, nick)
		ret = fmt.Sprintf("%s: Deleted nick %q", plugin, nick)
	}

	_wd.SaveFile(_cfgfile)
	return ret, nil
}

func safeArgs(num int, args []string) []string {
	alen := len(args)
	res := make([]string, num)
	for i := 0; i < num; i++ {
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

	if mtype(args[0], cmdList) {
		return ls(cmd.Channel, args[1], args[2]), nil
	} else if mtype(args[0], cmdAdd) {
		msg := strings.Join(cmd.Args[3:len(cmd.Args)], " ")
		return add(cmd.Channel, args[1], args[2], msg)
	} else if mtype(args[0], cmdDel) {
		return del(cmd.Channel, args[1], args[2])
	} else if mtype(args[0], cmdClear) {
		clear()
		retmsg = fmt.Sprintf("%s: DB cleared", plugin)
	} else if mtype(args[0], cmdReload) {
		reload()
		retmsg = fmt.Sprintf("%s: DB reloaded from disk", plugin)
	}

	return retmsg, nil
}
