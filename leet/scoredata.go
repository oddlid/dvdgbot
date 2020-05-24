package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	//"github.com/go-chat-bot/bot"
	"github.com/sirupsen/logrus"
)

type ScoreData struct {
	BotStart       time.Time           `json:"botstart"`
	Channels       map[string]*Channel `json:"channels"`
	saveInProgress bool
	calcInProgress bool
	l              *logrus.Entry
}

func newScoreData() *ScoreData {
	return &ScoreData{
		BotStart: time.Now(),
		Channels: make(map[string]*Channel),
	}
}

func (s *ScoreData) log() *logrus.Entry {
	if nil == s.l {
		return _log // pkg global
	}
	return s.l
}

func (s *ScoreData) load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, s)
}

func (s *ScoreData) loadFile(filename string) (*ScoreData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return s, err
	}
	defer file.Close()
	err = s.load(file)
	if err != nil {
		return s, err
	}
	s.log().WithField("filename", filename).Info("Leet stats (re)loaded from file")
	return s, nil
}

func (s *ScoreData) save(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(s, "", "\t")
	//jb, err := json.Marshal(s)
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (s *ScoreData) saveFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := s.save(file)
	if err != nil {
		return err
	}
	s.log().WithFields(logrus.Fields{
		"bytes":    n,
		"filename": filename,
	}).Info("File saved")
	return nil
}

func (s *ScoreData) scheduleSave(filename string, delayMinutes time.Duration) bool {
	if s.saveInProgress {
		return false
	}
	s.saveInProgress = true
	time.AfterFunc(delayMinutes*time.Minute, func() {
		err := s.saveFile(filename)
		if err != nil {
			s.log().WithError(err).Error("Scheduled save failed")
		}
		s.saveInProgress = false
	})
	return s.saveInProgress
}

func (s *ScoreData) calcAndPost(channel string) {
	c := s.get(channel)
	scoreMap := c.getScoresForRound()
	c.mergeScoresForRound(scoreMap)

	msg := fmt.Sprintf("New positive scores for %s:\n", time.Now().Format("2006-01-02"))
	fstr := getPadStrFmt(longestEntryLen(c.tmpNicks), ": %04d [+%02d] %s\n")

	getmsg := func(nick string, total, plus int) string {
		// The idea here is to print something extra if total points match any configured bonus value
		has, bc := _bonusConfigs.hasValue(total)
		if has {
			return fmt.Sprintf(fstr, nick, total, plus, bc.Greeting)
		}
		return fmt.Sprintf(fstr, nick, total, plus, "")
	}
	// This is the point where to calc random inspection and loss of points!

	for _, nick := range c.tmpNicks {
		msg += getmsg(nick, c.get(nick).getScore(), scoreMap[nick])
	}

	//_bot.SendMessage(
	//	bot.OutgoingMessage{
	//		channel,
	//		strings.TrimRight(msg, "\n"), // some servers post an empty line if present, so get rid of that
	//		&bot.User{},
	//		nil,
	//	},
	//)
	msgChan(channel, strings.TrimRight(msg, "\n"))

	// This is probably the best point to trigger an inspection and post the results
	// At any round, one contestant will be selected. But only a contestant, not someone who didn't participate this day
	// Selection is regular random of index between the the available ones in $scoreMap
	// So we let the happy news come first, and then we get mean and calculate the random victim for the day, and post
	// that with its updated/subtracted points value, but also if they were selected, but stayed clear.
	idx, tax := c.randomInspect()
	if idx > -1 {
		nick := c.tmpNicks[idx]
		user := c.get(nick)
		if tax > 0 {
			user.addScore(-tax)
			msg = fmt.Sprintf(
				"%s was randomly selected for taxation and lost %d points (now: %d points)",
				nick,
				tax,
				user.getScore(),
			)
		} else {
			msg = fmt.Sprintf("%s was randomly selected for taxation, but got off with a slap on the wrist ;)", nick)
		}
		msgChan(channel, msg)
	}

	c.clearNicksForRound() // clean up, before next round
}

func (s *ScoreData) scheduleCalcScore(channel string, delayMinutes time.Duration) bool {
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

func (s *ScoreData) get(channel string) *Channel {
	c, found := s.Channels[channel]
	if !found {
		c = &Channel{
			Users: make(map[string]*User),
			l:     s.log().WithField("channel", channel),
		}
		s.Channels[channel] = c
	}
	return c
}

func (s *ScoreData) rank(channel string) (KVList, error) {
	c := s.get(channel)
	if len(c.Users) == 0 {
		return nil, fmt.Errorf("ScoreData.rank(): No users with scores for channel %q", channel)
	}
	kvl := make(KVList, len(c.Users))
	i := 0
	for k, v := range c.Users {
		kvl[i] = KV{k, v.Points}
		i++
	}
	sort.Sort(sort.Reverse(kvl))
	return kvl, nil
}

func (s *ScoreData) stats(channel string) string {
	kvl, err := s.rank(channel)
	if err != nil {
		return err.Error()
	}
	fstr := getPadStrFmt(kvl.LongestKey(), ": %04d\n")
	str := fmt.Sprintf("Stats since %s:\n", s.BotStart.Format(time.RFC3339))
	for _, kv := range kvl {
		str += fmt.Sprintf(fstr, kv.Key, kv.Val)
	}
	return str
}

func (s *ScoreData) didTry(channel, nick string) bool {
	return s.get(channel).get(nick).hasTried()
}

func (s *ScoreData) tryScore(channel, nick string, t time.Time) (bool, string) {
	c := s.get(channel)
	points, tf := getScoreForEntry(t) // -1 or 0

	if TF_BEFORE == tf || TF_AFTER == tf {
		return false, ""
	}

	ts := fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	// This here is what makes it too hard to bother to get a custom greeting for a bonus, since
	// we add up the bonus points, but can't really add up greetings...
	//bonusPoints := _bonusConfigs.calc(fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond()))
	brs := _bonusConfigs.calc(fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond()))
	bonusPoints := brs.TotalBonus()
	_, userTotal := c.get(nick).score(points + bonusPoints)

	missTmpl := fmt.Sprintf("%s Too %s, sucker! %s: %d", ts, "%s", nick, userTotal)
	if bonusPoints > 0 {
		//missTmpl += fmt.Sprintf(" (but you got %d bonus points!)", bonusPoints)
		missTmpl += fmt.Sprintf(" (but: %s)", brs)
	}

	if TF_EARLY == tf {
		return true, fmt.Sprintf(missTmpl, "early")
	} else if TF_LATE == tf {
		return true, fmt.Sprintf(missTmpl, "late")
	}

	rank := c.addNickForRound(nick) // how many points is calculated from how many times this is called, later on

	ret := fmt.Sprintf("%s Whoop! %s: #%d", ts, nick, rank)
	if bonusPoints > 0 {
		//ret = fmt.Sprintf("%s (+%d points bonus!!!)", ret, bonusPoints)
		ret = fmt.Sprintf("%s (%s)", ret, brs)
	}

	return true, ret
}
