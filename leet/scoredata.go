package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

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
		l:        _log,
	}
}

func (s *ScoreData) isEmpty() bool {
	return len(s.Channels) == 0
}

func (s *ScoreData) log() *logrus.Entry {
	if nil == s.l {
		return _log
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

func (s *ScoreData) calcScore(c *Channel) string {
	scoreMap := c.getScoresForRound()
	var sb strings.Builder

	total := func(w io.Writer, val int) {
		fmt.Fprintf(w, ": %04d", val)
	}

	rank := func(w io.Writer, val int) {
		if 0 == val {
			return
		}
		fmt.Fprintf(w, " [Rank: +%02d]", val)
	}

	otax := func(w io.Writer, val int) {
		if 0 == val {
			return
		}
		fmt.Fprintf(w, " [Overshoot tax: -%d]", val)
	}

	tax := func(w io.Writer, val int) {
		if -1 == val {
			return
		}
		if 0 == val {
			fmt.Fprintf(w, " [Tax: Slap on the wrist ;)]")
			return
		}
		fmt.Fprintf(w, " [Tax: -%d]", val)
	}

	greeting := func(w io.Writer, total int) {
		has, bc := _bonusConfigs.hasValue(total)
		if !has {
			return
		}
		fmt.Fprintf(w, " - %s", bc.Greeting)
	}

	winner := func(w io.Writer, nick string) {
		isWinner := c.get(nick).isLocked()
		if !isWinner {
			return
		}
		fmt.Fprintf(w, " - Winner #%d!", c.getWinnerRank(nick)+1)
	}

	maxNickLen := c.Users.longestNickLen()

	genmsg := func(w io.Writer, nick string, tot, rnk, ostax, regtax int) {
		writePad(w, maxNickLen, nick)
		total(w, tot)
		rank(w, rnk)
		otax(w, ostax)
		tax(w, regtax)
		winner(w, nick)
		greeting(w, tot)
	}

	// If we are to use *OverShooters(), we need to remember to save other +- points to the users
	// first, as those funcs uses the users getScore() to determine how much to deduct or if the user should be on the list.
	// It might be better to just replicate it at the call site (here), if we need more flexibility.

	// generate header
	fmt.Fprintf(&sb, "Results for %s:\n", time.Now().Format("2006-01-02"))

	// taxNickIndex is the index of the taxed nick in c.tmpNicks
	taxNickIndex, taxVal := c.randomInspect() // taxNickIndex will be -2 if c.shouldInspect returns false because of weekday != rnd
	c.mergeScoresForRound(scoreMap)           // this needs to come before getOverShooters()
	osmap := c.getOverShooters(getTargetScore())
	// first we loop through the participants of this round that got on time and got points for that
	for idx, nick := range c.tmpNicks { // looping on tmpNicks will keep the sort order for most points
		// We need to compare each nick to entries in osmap, since we want to show the overshoot tax _either_ here, or
		// after this loop, but not both.
		u := c.get(nick)
		taxDeduction := -1
		if idx == taxNickIndex {
			taxDeduction = taxVal
		}
		rankPoints := scoreMap[nick] // this has been applied already, only for display purposes
		overshootTax := c.getOverShootTaxFor(getTargetScore(), u.getScore())
		// We now need to update the users points before we can get a greeting or mark as a winner
		if overshootTax > 0 {
			u.addScore(-overshootTax) // apply overshoot tax
			u.addTax(overshootTax)
		}
		if taxDeduction > 0 {
			u.addScore(-taxDeduction) // apply random tax
			u.addTax(taxDeduction)
		}
		// If the user is now at at total that matches target score, it needs to be marked as a winner, before we move on
		if getTargetScore() == u.getScore() {
			u.setLocked(true)
		}
		genmsg(&sb, nick, u.getScore(), rankPoints, overshootTax, taxDeduction)
		fmt.Fprintf(&sb, "\n")
	}
	// a user can be in osmap but not in tmpNicks if the user missed the time and got -1 for that, but also got a bonus
	// that made the total of those positive, and pushed the user to or over the limit
	for nick, user := range osmap {
		_, found := inStrSlice(c.tmpNicks, nick)
		if found {
			// If the overshooter is also a round contestant, we already dealt with it in the previous loop
			continue
		}
		overshootTax := c.getOverShootTaxFor(getTargetScore(), user.getScore())
		if overshootTax > 0 {
			user.addScore(-overshootTax)
		}
		if getTargetScore() == user.getScore() {
			user.setLocked(true)
		}
		genmsg(&sb, nick, user.getScore(), 0, overshootTax, -1)
		fmt.Fprintf(&sb, "\n")
	}

	c.clearNicksForRound() // clean up, before next round

	return sb.String()
}


func (s *ScoreData) scheduleCalcScore(c *Channel, delayMinutes time.Duration) bool {
	if s.calcInProgress {
		return false
	}
	s.calcInProgress = true
	time.AfterFunc(delayMinutes*time.Minute, func() {
		err := msgChan(c.Name, strings.TrimRight(s.calcScore(c), "\n"))
		if nil != err {
			s.log().WithField("func", "scheduleCalcScore").Error(err)
		}
		s.calcInProgress = false
	})
	return s.calcInProgress
}

func (s *ScoreData) get(channel string) *Channel {
	c, found := s.Channels[channel]
	if !found {
		c = &Channel{
			Name:  channel,
			Users: make(UserMap),
			l:     s.log().WithField("channel", channel),
		}
		s.Channels[channel] = c
		c.l.WithField("func", "get").Debug("Channel object created")
	}
	return c
}

func (s *ScoreData) stats(channel string) string {
	c := s.get(channel)
	var sb strings.Builder

	// This replaces the old func rank() that used KV/KVList
	us := c.Users.toSlice().sortByPointsDesc()

	// Since no changes to winner rank should happen during this method,
	// we pre-cache the list of winners here, and reimplement the functionality
	// of c.getWinnerRank, to speed up things a bit.
	ws := c.Users.filterByLocked(true).sortByLastEntryAsc()

	greeting := func(w io.Writer, total int) {
		has, bc := _bonusConfigs.hasValue(total)
		if !has {
			return
		}
		fmt.Fprintf(w, " - %s", bc.Greeting)
	}

	winner := func(w io.Writer, user *User) {
		isWinner := user.isLocked()
		if !isWinner {
			return
		}
		fmt.Fprintf(w, " - Winner #%d!", ws.getIndex(user.Nick)+1)
	}

	fstr := getPadStrFmt(
		c.Users.longestNickLen(),
		": %04d @ %s Best: %s Bonus: %03dx = %04d Tax: %03dx = -%04d",
	)

	fmt.Fprintf(&sb, "Stats since %s:\n", s.BotStart.Format(time.RFC3339))

	// It should be safe to access fields in user struct directly here without calling the methods
	// that lock, since we have guards otherwise that should prevent this method to be run in
	// parallell with anything.
	for _, u := range us {
		fmt.Fprintf(
			&sb,
			fstr,
			u.Nick,
			u.Points,
			getLongDate(u.getLastEntry()),
			getLongDate(u.getBestEntry()),
			u.getBonusTimes(),
			u.getBonusTotal(),
			u.getTaxTimes(),
			u.getTaxTotal(),
		)
		winner(&sb, u)
		greeting(&sb, u.Points)
		fmt.Fprintf(&sb, "\n")
	}

	return sb.String()
}

func (s *ScoreData) tryScore(c *Channel, u *User, t time.Time) (bool, string) {
	points, tf := getScoreForEntry(t) // -1 or 0

	// No points, not even minus if you're outside the timeframe
	if TF_BEFORE == tf || TF_AFTER == tf {
		return false, ""
	}

	ts := fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	brs := _bonusConfigs.calc(fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond()))
	bonusPoints := brs.TotalBonus()

	didScore, userTotal := u.score(points + bonusPoints, t)
	if !didScore {
		s.log().WithFields(logrus.Fields{
			"func":   "tryScore",
			"didScore": didScore,
		}).Error("It should not be possible to reach this branch")
		return false, fmt.Sprintf("%s: I'm retarded and made a logical error :'(", u.Nick)
	}

	missTmpl := fmt.Sprintf("%s Too %s, sucker! %s: %d", ts, "%s", u.Nick, userTotal)
	if bonusPoints > 0 {
		u.addBonus(bonusPoints)
		missTmpl += fmt.Sprintf(" (but: %s)", brs)
	}

	if TF_EARLY == tf {
		return true, fmt.Sprintf(missTmpl, "early")
	} else if TF_LATE == tf {
		return true, fmt.Sprintf(missTmpl, "late")
	}

	rank := c.addNickForRound(u.Nick) // how many points is calculated from how many times this is called, later on

	ret := fmt.Sprintf("%s Whoop! %s: #%d", ts, u.Nick, rank)
	if bonusPoints > 0 {
		ret = fmt.Sprintf("%s (%s)", ret, brs)
	}

	return true, ret
}
