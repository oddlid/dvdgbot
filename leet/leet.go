package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

const (
	DEF_HOUR   int    = 13
	DEF_MINUTE int    = 37
	SCORE_FILE string = "/tmp/leetbot_scores.json"
)

var (
	hour      int = DEF_HOUR
	minute    int = DEF_MINUTE
	scoreData *ScoreData
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
	Users       map[string]*User `json:"users"`
	singlePoint bool             // instead of "isFirst", so that it defaults to false and means less logic to write
}

type ScoreData struct {
	BotStart       time.Time           `json:"botstart"`
	Channels       map[string]*Channel `json:"channels"`
	saveInProgress bool
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

func NewScoreData() *ScoreData {
	return &ScoreData{
		BotStart: time.Now(),
		Channels: make(map[string]*Channel),
	}
}

func (u *User) Score(points int) (bool, int) {
	if u.didTry {
		return false, u.Points
	}
	u.Lock()
	u.Points += points
	u.didTry = true
	u.Unlock()
	time.AfterFunc(2*time.Minute, func() {
		u.Lock()
		u.didTry = false
		u.Unlock()
	})
	return true, u.Points
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

func (c *Channel) SetSinglePoint(val bool) {
	c.Lock()
	c.singlePoint = val
	c.Unlock()
}

func (c *Channel) GetScoreForEntry(t time.Time) (int, int) {
	if t.Hour() == hour {
		if t.Minute() == minute-1 {
			return -1, -1
		} else if t.Minute() == minute+1 {
			return -1, 1
		} else if t.Minute() == minute {
			if !c.singlePoint { // means first score within time frame
				c.SetSinglePoint(true) // next within time frame only gets 1 point
				// reset after 2 minutes
				time.AfterFunc(2*time.Minute, func() {
					c.SetSinglePoint(false)
				})
				return 2, 0
			} else {
				return 1, 0
			}
		}
	}
	return 0, 0
}

func (sd *ScoreData) Load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, sd)
}

func (s *ScoreData) LoadFile(filename string) *ScoreData {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("ScoreData.LoadFile(): Unable to load %q", filename)
		return s
	}
	defer file.Close()
	err = s.Load(file)
	if err != nil {
		log.Error(err)
		return NewScoreData()
	}
	log.Info("Leet stats (re)loaded from file")
	return s
}

func (sd *ScoreData) Save(w io.Writer) (int, error) {
	//jb, err := json.MarshalIndent(sd, "", "\t")
	jb, err := json.Marshal(sd)
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
	log.Infof("Saved %d bytes to %q", n, filename)
	return nil
}

func (s *ScoreData) ScheduleSave(filename string) bool {
	if s.saveInProgress {
		return false
	}
	s.saveInProgress = true
	time.AfterFunc(3*time.Minute, func() {
		err := s.SaveFile(filename)
		if err != nil {
			log.Errorf("Scheduled save failed: %s", err)
		}
		s.saveInProgress = false
	})
	return s.saveInProgress
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

	points, earlyOrLate := c.GetScoreForEntry(t)
	if points == 0 { // outside time frame
		return false, ""
	}

	didScore, userTotal := c.Get(nick).Score(points)
	if !didScore { // user tried again / spam
		return false, ""
	}

	ts := fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	if earlyOrLate == -1 {
		return true, fmt.Sprintf("%s Too early, sucker! %s: %d", ts, nick, userTotal)
	} else if earlyOrLate == 1 {
		return true, fmt.Sprintf("%s Too late, sucker! %s: %d", ts, nick, userTotal)
	} else if earlyOrLate == 0 {
		if points == 2 {
			return true, fmt.Sprintf("%s Double whoop! %s: %d", ts, nick, userTotal)
		} else if points == 1 {
			return true, fmt.Sprintf("%s Whoop! %s: %d", ts, nick, userTotal)
		}
	}

	// should not get here
	return false, ts
}

func withinTimeFrame(t time.Time) bool {
	if t.Hour() != hour {
		return false
	}
	curMin := t.Minute()
	if curMin != minute && curMin != minute-1 && curMin != minute+1 {
		return false
	}
	return true
}

func leet(cmd *bot.Cmd) (string, error) {
	t := time.Now() // save time as early as possible

	// handle arguments
	if len(cmd.Args) == 1 && cmd.Args[0] == "stats" {
		return scoreData.Stats(cmd.Channel), nil
	} else if len(cmd.Args) >= 1 {
		return fmt.Sprintf("Unrecognized argument: %q. Usage: !1337 [stats]", cmd.Args[0]), nil
	}

	// don't give a fuck outside accepted time frame
	if !withinTimeFrame(t) {
		return "", nil
	}

	// is the user spamming?
	if scoreData.DidTry(cmd.Channel, cmd.User.Nick) {
		//return "", fmt.Errorf("%s already tried within allowed timeframe", cmd.User.Nick)
		return fmt.Sprintf("%s: Stop spamming!", cmd.User.Nick), nil
	}

	success, msg := scoreData.TryScore(cmd.Channel, cmd.User.Nick, t)

	// at this point, data might have changed, and should be saved
	saveScheduled := scoreData.ScheduleSave(SCORE_FILE)
	if !saveScheduled {
		log.Debug("Ignoring redundant save scheduling")
	}

	if success {
		return msg, nil
	}

	// bogus
	return "", fmt.Errorf("Reached beyond logic...")
}

func pickupEnv() {
	h := os.Getenv("LEETBOT_HOUR")
	m := os.Getenv("LEETBOT_MINUTE")

	var err error
	if h != "" {
		hour, err = strconv.Atoi(h)
		if err != nil {
			hour = DEF_HOUR
		}
	}
	if m != "" {
		minute, err = strconv.Atoi(m)
		if err != nil {
			minute = DEF_MINUTE
		}
	}
}

func init() {
	scoreData = NewScoreData().LoadFile(SCORE_FILE)
	pickupEnv() // for minute/hour

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGUSR1:
				scoreData.LoadFile(SCORE_FILE)
			default:
				log.Info("Cought unhandled signal, ignoring")
			}
		}
	}()

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats]",
		leet)
}
