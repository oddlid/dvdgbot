package leet

import (
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
		return 0
	}
	maxTax := (float64(lowestTotal) / 100.0) * c.InspectionTax
	if maxTax < 0.0 {
		llog.WithField("maxTax", maxTax).Debug("Calculated tax points is negative, bailing out")
		return 0
	}
	return maxTax
}

func (c *Channel) shouldInspect() bool {
	if c.InspectAlways {
		return true
	}
	wd := int(time.Now().Weekday())
	rnd := rand.Intn(7)
	if wd != rnd {
		c.log().WithFields(logrus.Fields{
			"func":    "shouldInspect",
			"weekday": wd,
			"rnd":     rnd,
		}).Debug("Skipping inspection")
	}
	return wd == rnd
}

// Return index in c.tmpNicks and how many points minus, if selected, otherwise -1 (or -2) and 0
func (c *Channel) randomInspect() (int, int) {
	llog := c.log().WithField("func", "randomInspect")
	if !c.shouldInspect() {
		return -2, 0
	}
	maxTax := c.getMaxRoundTax()
	if maxTax < 1 {
		llog.WithField("maxTax", maxTax).Debug("Tax below 1, returning")
		return -1, 0
	}
	//llog.Debug("Got positive tax")
	return rand.Intn(len(c.tmpNicks)), rand.Intn(int(maxTax) + 1)
}
