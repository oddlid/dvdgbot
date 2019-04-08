package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

const (
	DEF_HOUR   int    = 13
	DEF_MINUTE int    = 37
	SCORE_FILE string = "/tmp/leetbot_scores.json"
	PLUGIN     string = "LeetBot"
)

type TimeCode int

// Constants for signalling how much off timeframe
const (
	TF_BEFORE TimeCode = iota // more than a minute before
	TF_EARLY                  // less than a minute before
	TF_ONTIME                 // within correct minute
	TF_LATE                   // less than a minute late
	TF_AFTER                  // more than a minute late
)

var (
	_hour      int = DEF_HOUR
	_minute    int = DEF_MINUTE
	_scoreData *ScoreData
	_bot       *bot.Bot
)

type KV struct {
	Key string
	Val int
}

type KVList []KV

type User struct {
	sync.RWMutex
	Points int `json:"score"`
	didTry bool
}

type Channel struct {
	sync.RWMutex
	Users    map[string]*User `json:"users"`
	tmpNicks []string         // used for storing who participated in a specific round. Reset after calculation.
}

type ScoreData struct {
	BotStart       time.Time           `json:"botstart"`
	Channels       map[string]*Channel `json:"channels"`
	saveInProgress bool
	calcInProgress bool
}

func (kl KVList) Len() int {
	return len(kl)
}

func (kl KVList) Less(i, j int) bool {
	return kl[i].Val < kl[j].Val
}

func (kl KVList) Swap(i, j int) {
	kl[i], kl[j] = kl[j], kl[i]
}

func SetParentBot(b *bot.Bot) {
	_bot = b
}

func NewScoreData() *ScoreData {
	return &ScoreData{
		BotStart: time.Now(),
		Channels: make(map[string]*Channel),
	}
}

// Helper func pb = "prefixed by". Returns true if all chars before given
// position are what's given as "char" argument.
func pb(str string, char rune, pos int) bool {
	for i, r := range str {
		if r != char {
			return false
		}
		if i >= pos-1 {
			break
		}
	}
	return true
}

// bonus() returns extra points if the timestamp has certain patterns
func bonus(t time.Time) int {
	// We use the given hour and minute for point patterns.
	// The farther to the right the pattern occurs, the more points.
	// So, if hour = 13, minute = 37, we'd get something like this:
	// 13:37:13:37xxxxx = +1 point
	// 13:37:01:337xxxx = +2 points
	// 13:37:00:1337xxx = +3 points
	// 13:37:00:01337xx = +4 points
	// 13:37:00:001337x = +5 points
	// 13:37:00:0001337 = +6 points
	// Tighter than 3x0 is very unlikely anyone will ever get.
	// Still, we'll just calculate the bonus by using substring index +1.

	bonus := 0
	ts := fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond())
	sstr := fmt.Sprintf("%02d%02d", _hour, _minute)
	idx := strings.Index(ts, sstr)

	if idx > -1 {
		if idx > 0 {
			// make sure it's only 0's before the match
			if pb(ts, '0', idx) {
				bonus = idx + 1
			} else { // still give a small bonus if match, but not prefixed with 0's
				bonus = 1
			}
		} else {
			bonus = 1
		}
	}

	return bonus
}

func (u *User) AddScore(points int) int {
	u.Lock()
	defer u.Unlock()
	u.Points += points
	return u.Points
}

func (u *User) Score(points int) (bool, int) {
	if u.didTry {
		return false, u.Points
	}
	u.AddScore(points)
	u.Lock()
	u.didTry = true
	total := u.Points
	u.Unlock()
	// Reset didTry after 2 minutes
	// This should create a "loophole" so that if a user posts too early and gets -1,
	// they could manage to get another -1 by being too late as well :D
	time.AfterFunc(2*time.Minute, func() {
		u.Lock()
		u.didTry = false
		u.Unlock()
	})
	return true, total
}

func (c *Channel) Get(nick string) *User {
	c.RLock()
	user, found := c.Users[nick]
	c.RUnlock()
	if !found {
		user = &User{}
		c.Lock()
		c.Users[nick] = user
		c.Unlock()
	}
	return user
}

func (c *Channel) HasPendingScores() bool {
	c.RLock()
	defer c.RUnlock()
	return c.tmpNicks != nil && len(c.tmpNicks) > 0
}

func (c *Channel) AddNickForRound(nick string) int {
	// first in gets the most points, last the least
	c.Lock()
	defer c.Unlock()
	c.tmpNicks = append(c.tmpNicks, nick)
	return len(c.tmpNicks) // returns first place, second place etc
}

func (c *Channel) ClearNicksForRound() {
	c.Lock()
	c.tmpNicks = nil
	c.Unlock()
}

// GetScoresForRound returns a map of nicks with the scores for this round
func (c *Channel) GetScoresForRound() map[string]int {
	if c.tmpNicks == nil || len(c.tmpNicks) == 0 {
		return nil
	}
	maxScore := len(c.tmpNicks)
	nickMap := make(map[string]int)
	c.Lock()
	for i := range c.tmpNicks {
		nickMap[c.tmpNicks[i]] = maxScore - i
	}
	c.Unlock()

	return nickMap
}

func (c *Channel) MergeScoresForRound(newScores map[string]int) {
	for nick := range newScores {
		c.Get(nick).AddScore(newScores[nick])
	}
}

func (c *Channel) GetScoreForEntry(t time.Time) (int, TimeCode) {
	// Don't know why I made this a method on Channel, as it does not use it
	var points int
	tf := timeFrame(t)

	if TF_EARLY == tf || TF_LATE == tf {
		points = -1
	} else {
		points = 0 // will be set later if on time
	}

	return points, tf
}

func (s *ScoreData) Load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, s)
}

func (s *ScoreData) LoadFile(filename string) *ScoreData {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("%s: ScoreData.LoadFile() Error: %q", PLUGIN, err.Error())
		return s
	}
	defer file.Close()
	err = s.Load(file)
	if err != nil {
		log.Error(err)
		return NewScoreData()
	}
	log.Infof("%s: Leet stats (re)loaded from file %q", PLUGIN, filename)
	return s
}

func (s *ScoreData) Save(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(s, "", "\t")
	//jb, err := json.Marshal(s)
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (s *ScoreData) SaveFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := s.Save(file)
	if err != nil {
		return err
	}
	log.Infof("%s: Saved %d bytes to %q", PLUGIN, n, filename)
	return nil
}

func (s *ScoreData) ScheduleSave(filename string, delayMinutes time.Duration) bool {
	if s.saveInProgress {
		return false
	}
	s.saveInProgress = true
	time.AfterFunc(delayMinutes*time.Minute, func() {
		err := s.SaveFile(filename)
		if err != nil {
			log.Errorf("%s: Scheduled save failed: %s", PLUGIN, err.Error())
		}
		s.saveInProgress = false
	})
	return s.saveInProgress
}

func (s *ScoreData) calcAndPost(channel string) {
	c := s.Get(channel)
	scoreMap := c.GetScoresForRound()
	c.MergeScoresForRound(scoreMap)

	// first loop for getting max nicklen
	nick_maxlen := 0
	for i := range c.tmpNicks {
		nlen := len(c.tmpNicks[i])
		if nlen > nick_maxlen {
			nick_maxlen = nlen
		}
	}
	msg := fmt.Sprintf("New positive scores for %s:\n", time.Now().Format("2006-01-02"))
	fstr := fmt.Sprintf("%s%d%s", "%-", nick_maxlen, "s : %04d [+%02d]\n")
	for _, nick := range c.tmpNicks {
		msg += fmt.Sprintf(fstr, nick, c.Get(nick).Points, scoreMap[nick])
	}

	_bot.SendMessage(
		channel,
		msg,
		&bot.User{}, // trying empty user struct, might be enough
	)

	c.ClearNicksForRound() // clean up, before next round
}

func (s *ScoreData) ScheduleCalcScore(channel string, delayMinutes time.Duration) bool {
	if s.calcInProgress {
		return false
	}
	s.calcInProgress = true
	time.AfterFunc(delayMinutes*time.Minute, func() {
		s.calcAndPost(channel)
		s.calcInProgress = false
	})
	return s.calcInProgress
}

func (s *ScoreData) Get(channel string) *Channel {
	c, found := s.Channels[channel]
	if !found {
		c = &Channel{
			Users: make(map[string]*User),
		}
		s.Channels[channel] = c
	}
	return c
}

func (s *ScoreData) Rank(channel string) (KVList, int, error) {
	c := s.Get(channel)
	if len(c.Users) == 0 {
		return nil, 0, fmt.Errorf("ScoreData.Rank(): No users with scores for channel %q", channel)
	}
	kl := make(KVList, len(c.Users))
	i := 0
	nick_maxlen := 0
	for k, v := range c.Users {
		kl[i] = KV{k, v.Points}
		i++
		nlen := len(k)
		if nlen > nick_maxlen {
			nick_maxlen = nlen
		}
	}
	sort.Sort(sort.Reverse(kl))
	return kl, nick_maxlen, nil
}

func (s *ScoreData) Stats(channel string) string {
	kl, max_nicklen, err := s.Rank(channel)
	if err != nil {
		return err.Error()
	}
	fstr := fmt.Sprintf("%s%d%s", "%-", max_nicklen, "s : %04d\n")
	str := fmt.Sprintf("Stats since %s:\n", s.BotStart.Format(time.RFC3339))
	for _, kv := range kl {
		str += fmt.Sprintf(fstr, kv.Key, kv.Val)
	}
	return str
}

func (s *ScoreData) DidTry(channel, nick string) bool {
	return s.Get(channel).Get(nick).didTry
}

func (s *ScoreData) TryScore(channel, nick string, t time.Time) (bool, string) {
	c := s.Get(channel)
	points, tf := c.GetScoreForEntry(t)

	if TF_BEFORE == tf || TF_AFTER == tf {
		return false, ""
	}

	ts := fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	bonusPoints := bonus(t)
	_, userTotal := c.Get(nick).Score(points + bonusPoints)

	missTmpl := fmt.Sprintf("%s Too %s, sucker! %s: %d", ts, "%s", nick, userTotal)
	if bonusPoints > 0 {
		missTmpl += fmt.Sprintf(" (but you got %d bonus points!)", bonusPoints)
	}

	if TF_EARLY == tf {
		//return true, fmt.Sprintf("%s Too early, sucker! %s: %d", ts, nick, userTotal)
		return true, fmt.Sprintf(missTmpl, "early")
	} else if TF_LATE == tf {
		//return true, fmt.Sprintf("%s Too late, sucker! %s: %d", ts, nick, userTotal)
		return true, fmt.Sprintf(missTmpl, "late")
	}

	rank := c.AddNickForRound(nick) // how many points is calculated from how many times this is called, later on

	ret := fmt.Sprintf("%s Whoop! %s: #%d", ts, nick, rank)
	if bonusPoints > 0 {
		ret = fmt.Sprintf("%s (+%d points bonus!!!)", ret, bonusPoints)
	}

	return true, ret
}

func timeFrame(t time.Time) TimeCode {
	th := t.Hour()
	if th < _hour {
		return TF_BEFORE
	} else if th > _hour {
		return TF_AFTER
	}
	// now we know we're within the correct hour
	tm := t.Minute()
	if tm < _minute-1 {
		return TF_BEFORE
	} else if tm > _minute+1 {
		return TF_AFTER
	} else if tm == _minute-1 {
		return TF_EARLY
	} else if tm == _minute+1 {
		return TF_LATE
	}
	return TF_ONTIME
}

func withinTimeFrame(t time.Time) (bool, TimeCode) {
	tf := timeFrame(t)
	if tf == TF_EARLY || tf == TF_ONTIME || tf == TF_LATE {
		return true, tf
	}
	return false, tf
}

func leet(cmd *bot.Cmd) (string, error) {
	t := time.Now() // save time as early as possible

	// handle arguments
	alen := len(cmd.Args)
	if alen == 1 && "stats" == cmd.Args[0] {
		if _scoreData.calcInProgress {
			return "Stats are calculating. Try again in a couple of minutes.", nil
		} else {
			return _scoreData.Stats(cmd.Channel), nil
		}
	} else if alen == 1 && "reload" == cmd.Args[0] {
		if !_scoreData.saveInProgress {
			_scoreData.LoadFile(SCORE_FILE)
			return "Score data reloaded from file", nil
		} else {
			return "A scheduled save is in progress. Will not reload right now.", nil
		}
	} else if alen >= 1 {
		return fmt.Sprintf("Unrecognized argument: %q. Usage: !1337 [stats|reload]", cmd.Args[0]), nil
	}

	// don't give a fuck outside accepted time frame
	inTimeFrame, tf := withinTimeFrame(t)
	if !inTimeFrame {
		return "", nil
	}

	// is the user spamming?
	if _scoreData.DidTry(cmd.Channel, cmd.User.Nick) {
		return fmt.Sprintf("%s: Stop spamming!", cmd.User.Nick), nil
	}

	success, msg := _scoreData.TryScore(cmd.Channel, cmd.User.Nick, t)

	// at this point, data might have changed, and should be saved
	var delayMinutes time.Duration
	if TF_EARLY == tf {
		delayMinutes = 3
	} else if TF_ONTIME == tf {
		delayMinutes = 2
	} else if TF_LATE == tf {
		delayMinutes = 1
	}

	if success && !_scoreData.saveInProgress {
		_scoreData.ScheduleSave(SCORE_FILE, delayMinutes+1)
	}

	if !_scoreData.calcInProgress && _scoreData.Get(cmd.Channel).HasPendingScores() {
		_scoreData.ScheduleCalcScore(cmd.Channel, delayMinutes)
	}

	if success {
		return msg, nil
	}

	// bogus
	return "", fmt.Errorf("%s: Reached beyond logic...", PLUGIN)
}

func pickupEnv() {
	h := os.Getenv("LEETBOT_HOUR")
	m := os.Getenv("LEETBOT_MINUTE")

	var err error
	if h != "" {
		_hour, err = strconv.Atoi(h)
		if err != nil {
			_hour = DEF_HOUR
		}
	}
	if m != "" {
		_minute, err = strconv.Atoi(m)
		if err != nil {
			_minute = DEF_MINUTE
		}
	}
}

func init() {
	_scoreData = NewScoreData().LoadFile(SCORE_FILE)
	pickupEnv() // for minute/hour

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
}
