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
	Name          string   `json:"channel_name,omitempty"` // we need to duplicate this from the parent map key, so that the instance knows its own name
	Users         UserMap  `json:"users"`                  // string key is nick
	InspectionTax float64  `json:"inspection_tax"`         // percentage, but no check if outside of 0-100
	InspectAlways bool     `json:"inspect_always"`         // if false, only inspect if random value between 0 and 6 matches current weekday
	TaxLoners     bool     `json:"tax_loners"`             // If to inspect and tax when only one contestant in a round
	PostTaxFail   bool     `json:"post_tax_fail"`          // If to post to channel why taxation does NOT happen
	OvershootTax  int      `json:"overshoot_tax"`          // interval for how much to deduct if user scores past target
	tmpNicks      []string // used for storing who participated in a specific round. Reset after calculation.
	l             *logrus.Entry
}

func (c *Channel) log() *logrus.Entry {
	if nil == c.l {
		// Instead of setting c.l here, we just return _log, so that we may set c.l later in name() instead, with more fields
		return _log
	}
	return c.l
}

func (c *Channel) get(nick string) *User {
	c.RLock()
	user, found := c.Users[nick]
	c.RUnlock()
	if !found {
		user = &User{
			Nick: nick,
			l:    c.log().WithField("user", nick),
		}
		c.Lock()
		c.Users[nick] = user
		c.Unlock()
	}
	return user
}

func (c *Channel) name() (string, error) {
	key := "channel"
	if c.Name != "" {
		if nil == c.l {
			c.l = _log.WithField(key, c.Name)
		}
		return c.Name, nil
	}
	if nil != c.l {
		entry, found := c.l.Data[key] // type Fields map[string]interface{}
		if !found {
			return "", fmt.Errorf("No logrus field with key: %q", key)
		}
		c.Name = fmt.Sprintf("%v", entry) // set name for later
		return c.Name, nil
	}
	return "", fmt.Errorf("Unable to resolve channel name")
}

func (c *Channel) nickList() []string {
	nicks := make([]string, 0, len(c.Users))
	for k := range c.Users {
		nicks = append(nicks, k)
	}
	return nicks
}

func (c *Channel) postTaxFail(msg string) error {
	llog := c.log().WithField("func", "postTaxFail")
	if !c.PostTaxFail {
		str := "Configured to NOT post tax fail"
		llog.Debug(str)
		return fmt.Errorf(str)
	}

	cname, err := c.name()
	if nil != err {
		llog.Error(err)
		return err
	}

	err = msgChan(cname, msg)
	if nil != err {
		llog.Error(err)
	}

	return err
}

//func (c *Channel) maxPoints() (nick string, res int) {
//	for k, v := range c.Users {
//		if v.Points > res {
//			res = v.Points
//			nick = k
//		}
//	}
//	return
//}

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

// removeNickFromRound returns true if nick was found and deleted, false otherwise
// Not in use now, but keeping it commented for reference, in case we need it later.
//func (c *Channel) removeNickFromRound(nick string) bool {
//	if nil == c.tmpNicks || len(c.tmpNicks) == 0 {
//		return false
//	}
//
//	nickIdx := -1
//	for idx := range c.tmpNicks {
//		if nick == c.tmpNicks[idx] {
//			nickIdx = idx
//			break
//		}
//	}
//
//	if -1 == nickIdx {
//		return false
//	}
//
//	// Use slow version that maintains order, from https://yourbasic.org/golang/delete-element-slice/
//	numNicks := len(c.tmpNicks)
//	copy(c.tmpNicks[nickIdx:], c.tmpNicks[nickIdx+1:])
//	c.tmpNicks[numNicks-1] = ""
//	c.tmpNicks = c.tmpNicks[:numNicks-1]
//
//	return true
//}

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

// getOverShooters will return both those who got exactly to the target point sum,
// and those that got past it.
// Since it will be possible to miss so one's not included in tmpNicks, but still get a bonus
// that takes you past the limit, we need to check all users here.
func (c *Channel) getOverShooters(limit int) UserMap {
	ret := make(UserMap)
	c.RLock()
	for nick, user := range c.Users {
		if user.getScore() >= limit {
			ret[nick] = user
		}
	}
	c.RUnlock()
	return ret
}

func (c *Channel) getOverShootTaxFor(limit, points int) int {
	// Setting OvershootTax to 0 or below should disable taxation
	if c.OvershootTax <= 0 {
		return 0
	}

	if limit == points {
		return 0
	}

	deduction := 0
	for points-deduction >= limit {
		deduction += c.OvershootTax
	}

	return deduction
}

// TODO: Rewrite tests, as they are the only ones using this method, so we can remove this
func (c *Channel) punishOverShooters(limit int, umap UserMap) UserMap {
	c.Lock()
	for _, user := range umap {
		tax := c.getOverShootTaxFor(limit, user.getScore())
		user.addScore(-tax)
	}
	c.Unlock()
	return umap
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

// 2021-03-09 22:14
// We need getters and setters for InspectAlways, as when running tests with ./... they fail because
// of concurrent access to this setting.
// Update: Seems that wasn't the culprit after all. It was forgetting to call c.cleaNicksForRound()...
// But anyways, this is better anyways, so keeping it.

func (c *Channel) setInspectAlways(doInspect bool) {
	c.Lock()
	c.InspectAlways = doInspect
	c.Unlock()
}

func (c *Channel) getInspectAlways() bool {
	c.RLock()
	defer c.RUnlock()
	return c.InspectAlways
}

func (c *Channel) setTaxLoners(doTax bool) {
	c.Lock()
	c.TaxLoners = doTax
	c.Unlock()
}

func (c *Channel) getTaxLoners() bool {
	c.RLock()
	defer c.RUnlock()
	return c.TaxLoners
}

func (c *Channel) shouldInspect() bool {
	llog := c.log().WithFields(logrus.Fields{
		"func": "shouldInspect",
	})
	// Having this check before the next will override TaxLoners
	if c.getInspectAlways() {
		llog.Debug("Configured to always run inspection")
		return true
	}
	// We could have something like this to only tax when more than 1 contestant
	if nil == c.tmpNicks || (!c.getTaxLoners() && len(c.tmpNicks) < 2) {
		llog.Debug("Configured to NOT tax loners")
		return false
	}

	wd := int(time.Now().Weekday())
	rnd := rand.Intn(7) // 7 for number of days in week
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
		nickIndex = -2 // unique "error" value indicating where this func bailed out
		return
	}
	maxTax := c.getMaxRoundTax()
	if maxTax < 1 { // I don't think we've ever reached this section irl
		llog.WithField("maxTax", maxTax).Debug("Tax below 1, returning")
		err := c.postTaxFail(fmt.Sprintf("No tax today. Calculated tax was: %f", maxTax))
		if nil != err {
			llog.Error(err)
		}
		nickIndex = -1
		return
	}

	nickIndex = rand.Intn(len(c.tmpNicks))
	tax = rand.Intn(int(maxTax) + 1)
	return
}

// Calling this repeatedly might be inefficient and wasteful.
// Might be better to implement a variant at the call site.
// Update: Benchmarks showed this to be over 20x slower when called
// repeatedly for a list of 7 "winners", rather than first getting
// the filtered and sorted list, and then running getIndex with the
// list cached. So yes, very wasteful. But we still need it some places,
// as it would otherwise be too cumbersome, like in ScoreData.calcScore.
// But in that method, speed doesn't matter that much, as it happens after
// everyone is done trying to score as fast as possible.
func (c *Channel) getWinnerRank(nick string) int {
	return c.Users.filterByLocked(true).sortByLastEntryAsc().getIndex(nick)
}
