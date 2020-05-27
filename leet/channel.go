package leet

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Channel struct {
	sync.RWMutex
	Users         map[string]*User `json:"users"`
	InspectionTax float64          `json:"inspection_tax"` // percentage, but no check if outside of 0-100
	InspectAlways bool             `json:"inspect_always"` // if false, only inspect if random value between 0 and 6 matches current weekday
	TaxLoners     bool             `json:"tax_loners"`     // If to inspect and tax when only one contestant in a round
	PostTaxFail   bool             `json:"post_tax_fail"`  // If to post to channel why taxation does NOT happen
	tmpNicks      []string         // used for storing who participated in a specific round. Reset after calculation.
	l             *logrus.Entry
}

func (c *Channel) log() *logrus.Entry {
	if nil == c.l {
		return _log // pkg global
	}
	return c.l
}

func (c *Channel) get(nick string) *User {
	c.RLock()
	user, found := c.Users[nick]
	c.RUnlock()
	if !found {
		user = &User{
			l: c.log().WithField("user", nick),
		}
		c.Lock()
		c.Users[nick] = user
		c.Unlock()
	}
	return user
}

// Could have saved the name of the channel in the struct, but since we already have it
// in the logrus.Entry, we can pull it from that without adding to the struct
func (c *Channel) name() (string, error) {
	// can't rely on the log() method here, as that will return an Entry without the channel name
	// if the channels Entry is not set
	if nil == c.l {
		return "", fmt.Errorf("log is nil, can't derive channel name")
	}
	entry, found := c.l.Data["channel"] // type Fields map[string]interface{}
	if !found {
		return "", fmt.Errorf("no logrus field with key \"channel\"")
	}
	return fmt.Sprintf("%v", entry), nil
}

func (c *Channel) post(msg string) error {
	llog := c.log().WithField("func", "post")
	if "" == msg {
		str := "Refusing to post empty message"
		llog.Error(str)
		return fmt.Errorf(str)
	}
	cname, err := c.name()
	if nil != err {
		llog.Error(err)
		return err
	}
	llog.WithFields(logrus.Fields{
		"message": msg,
		"channel": cname,
	}).Debug("Delegating channel post to parent bot")

	return msgChan(cname, msg) // delegate to parent
}

func (c *Channel) postTaxFail(msg string) error {
	llog := c.log().WithField("func", "postTaxFail")
	if !c.PostTaxFail {
		str := "Configured to NOT post tax fail"
		llog.Debug(str)
		return fmt.Errorf(str)
	}
	err := c.post(msg)
	if nil != err {
		llog.Error(err)
	}
	return err
}

func (c *Channel) maxPoints() (nick string, res int) {
	for k, v := range c.Users {
		if v.Points > res {
			res = v.Points
			nick = k
		}
	}
	return
}

func (c *Channel) hasPendingScores() bool {
	c.RLock()
	defer c.RUnlock()
	return c.tmpNicks != nil && len(c.tmpNicks) > 0
}

func (c *Channel) addNickForRound(nick string) int {
	// first in gets the most points, last the least
	c.Lock()
	defer c.Unlock()
	c.tmpNicks = append(c.tmpNicks, nick)
	return len(c.tmpNicks) // returns first place, second place etc
}

func (c *Channel) clearNicksForRound() {
	c.Lock()
	c.tmpNicks = nil
	c.Unlock()
}

// GetScoresForRound returns a map of nicks with the scores for this round
func (c *Channel) getScoresForRound() map[string]int {
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

func (c *Channel) mergeScoresForRound(newScores map[string]int) {
	for nick := range newScores {
		c.get(nick).addScore(newScores[nick])
	}
}

// Find the lowest total points for the users who participated in the current round
func (c *Channel) getLowestTotalInRound() int {
	if nil == c.tmpNicks || len(c.tmpNicks) == 0 {
		c.log().WithFields(logrus.Fields{
			"func": "getLowestTotalInRound",
		}).Debug("tmpNicks is empty, bailing out")
		return 0
	}
	lowestTotal := c.get(c.tmpNicks[0]).getScore()
	for _, nick := range c.tmpNicks {
		score := c.get(nick).getScore()
		if score < lowestTotal {
			lowestTotal = score
		}
	}
	return lowestTotal
}

func (c *Channel) getMaxRoundTax() float64 {
	llog := c.log().WithField("func", "getMaxRoundTax")

	if c.InspectionTax <= 0.0 { // use as a way to disable this functionality
		llog.WithField("InspectionTax", c.InspectionTax).Debug("Negative or zero InspectionTax, bailing out")
		return 0
	}
	lowestTotal := c.getLowestTotalInRound()
	if lowestTotal < 1 {
		llog.WithField("lowestTotal", lowestTotal).Debug("Lowest total is below 1, bailing out")
		err := c.postTaxFail("No tax today, as we have a participant with less than 1 points")
		if nil != err {
			llog.Error(err)
		}
		return 0
	}
	maxTax := (float64(lowestTotal) / 100.0) * c.InspectionTax
	if maxTax < 0.0 {
		llog.WithField("maxTax", maxTax).Debug("Calculated tax points is negative, bailing out")
		err := c.postTaxFail(fmt.Sprintf("No tax today: Lowest total = %d, Percent = %f - which amounts to %f", lowestTotal, c.InspectionTax, maxTax))
		if nil != err {
			llog.Error(err)
		}
		return 0
	}
	return maxTax
}

func (c *Channel) shouldInspect() bool {
	llog := c.log().WithFields(logrus.Fields{
		"func": "shouldInspect",
	})
	// Having this check before the next will override TaxLoners
	if c.InspectAlways {
		llog.Debug("Configured to always run inspection")
		return true
	}
	// We could have something like this to only tax when more than 1 contestant
	if nil == c.tmpNicks || (!c.TaxLoners && len(c.tmpNicks) < 2) {
		llog.Debug("Configured to NOT tax loners")
		return false
	}

	wd := int(time.Now().Weekday())
	rnd := rand.Intn(7)
	doInspect := wd == rnd
	llog.WithFields(logrus.Fields{
		"weekday": wd,
		"rnd":     rnd,
		"inspect": doInspect,
	}).Debug("To inspect or not...")

	if !doInspect {
		err := c.postTaxFail(fmt.Sprintf("No tax today :) Weekday = %d, random = %d", wd, rnd))
		if nil != err {
			llog.Error(err)
		}
	}

	return doInspect
}

// Return index in c.tmpNicks and how many points minus, if selected, otherwise -1 (or -2) and 0
func (c *Channel) randomInspect() (nickIndex, tax int) {
	llog := c.log().WithField("func", "randomInspect")
	if !c.shouldInspect() {
		nickIndex = -2
		return
	}
	maxTax := c.getMaxRoundTax()
	if maxTax < 1 {
		llog.WithField("maxTax", maxTax).Debug("Tax below 1, returning")
		c.postTaxFail(fmt.Sprintf("No tax today. Calculated tax was: %f", maxTax))
		nickIndex = -1
		return
	}

	nickIndex = rand.Intn(len(c.tmpNicks))
	tax = rand.Intn(int(maxTax) + 1)
	return
}
