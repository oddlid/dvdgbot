package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

type ScoreData struct {
	BotStart       time.Time           `json:"botstart"`
	Channels       map[string]*Channel `json:"channels"`
	saveInProgress bool
	calcInProgress bool
}

func NewScoreData() *ScoreData {
	return &ScoreData{
		BotStart: time.Now(),
		Channels: make(map[string]*Channel),
	}
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
	fstr := fmt.Sprintf("%s%d%s", "%-", nick_maxlen, "s : %04d [+%02d]%s\n")

	getmsg := func(nick string, total, plus int) string {
		// The idea here is to print something extra if total points match any configured bonus value
		xtra := ""
		if _bonusConfigs.HasValue(total) {
			xtra = " Hail Satan \m/"
		}
		return fmt.Sprintf(fstr, nick, total, plus, xtra)
	}

	for _, nick := range c.tmpNicks {
		//msg += fmt.Sprintf(fstr, nick, c.Get(nick).Points, scoreMap[nick])
		msg += getmsg(nick, c.Get(nick).Points, scoreMap[nick])
	}

	_bot.SendMessage(
		bot.OutgoingMessage{
			channel,
			msg,
			&bot.User{},
			nil,
		},
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

	bonusPoints := _bonusConfigs.Calc(fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond()))
	_, userTotal := c.Get(nick).Score(points + bonusPoints)

	missTmpl := fmt.Sprintf("%s Too %s, sucker! %s: %d", ts, "%s", nick, userTotal)
	if bonusPoints > 0 {
		missTmpl += fmt.Sprintf(" (but you got %d bonus points!)", bonusPoints)
	}

	if TF_EARLY == tf {
		return true, fmt.Sprintf(missTmpl, "early")
	} else if TF_LATE == tf {
		return true, fmt.Sprintf(missTmpl, "late")
	}

	rank := c.AddNickForRound(nick) // how many points is calculated from how many times this is called, later on

	ret := fmt.Sprintf("%s Whoop! %s: #%d", ts, nick, rank)
	if bonusPoints > 0 {
		ret = fmt.Sprintf("%s (+%d points bonus!!!)", ret, bonusPoints)
	}

	return true, ret
}
