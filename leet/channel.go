package leet

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Channel struct {
	l             zerolog.Logger
	Users         UserMap  `json:"users"`                  // string key is nick
	Name          string   `json:"channel_name,omitempty"` // we need to duplicate this from the parent map key, so that the instance knows its own name
	tmpNicks      []string // used for storing who participated in a specific round. Reset after calculation.
	InspectionTax float64  `json:"inspection_tax"` // percentage, but no check if outside of 0-100
	OvershootTax  int      `json:"overshoot_tax"`  // interval for how much to deduct if user scores past target
	mu            sync.RWMutex
	InspectAlways bool `json:"inspect_always"` // if false, only inspect if random value between 0 and 6 matches current weekday
	TaxLoners     bool `json:"tax_loners"`     // If to inspect and tax when only one contestant in a round
	PostTaxFail   bool `json:"post_tax_fail"`  // If to post to channel why taxation does NOT happen
}

func (c *Channel) get(nick string) *User {
	c.mu.RLock()
	user, found := c.Users[nick]
	c.mu.RUnlock()
	if !found {
		user = &User{
			Nick: nick,
			l:    c.l.With().Str("user", nick).Logger(),
		}
		c.mu.Lock()
		c.Users[nick] = user
		c.mu.Unlock()
	}
	return user
}

func (c *Channel) nickList() []string {
	nicks := make([]string, 0, len(c.Users))
	for k := range c.Users {
		nicks = append(nicks, k)
	}
	return nicks
}

func (c *Channel) postTaxFail(msg string) error {
	if !c.PostTaxFail {
		return fmt.Errorf("configured to NOT post tax fail")
	}

	return msgChan(c.Name, msg)
}

func (c *Channel) hasPendingScores() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tmpNicks != nil && len(c.tmpNicks) > 0
}

func (c *Channel) addNickForRound(nick string) int {
	// first in gets the most points, last the least
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tmpNicks = append(c.tmpNicks, nick)
	return len(c.tmpNicks) // returns first place, second place etc
}

func (c *Channel) clearNicksForRound() {
	c.mu.Lock()
	c.tmpNicks = nil
	c.mu.Unlock()
}

// GetScoresForRound returns a map of nicks with the scores for this round
func (c *Channel) getScoresForRound() map[string]int {
	maxScore := len(c.tmpNicks)
	if maxScore == 0 {
		return nil
	}
	nickMap := make(map[string]int)
	c.mu.Lock()
	for i := range c.tmpNicks {
		nickMap[c.tmpNicks[i]] = maxScore - i
	}
	c.mu.Unlock()

	return nickMap
}

func (c *Channel) mergeScoresForRound(newScores map[string]int) {
	for nick := range newScores {
		c.get(nick).addScore(newScores[nick])
	}
}

// Find the lowest total points for the users who participated in the current round
func (c *Channel) getLowestTotalInRound() int {
	if c.tmpNicks == nil || len(c.tmpNicks) == 0 {
		c.l.Debug().
			Str("func", "getLowestTotalInRound").
			Msg("tmpNicks is empty, bailing out")
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
	c.mu.RLock()
	for nick, user := range c.Users {
		if user.getScore() >= limit {
			ret[nick] = user
		}
	}
	c.mu.RUnlock()
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
	c.mu.Lock()
	for _, user := range umap {
		tax := c.getOverShootTaxFor(limit, user.getScore())
		user.addScore(-tax)
	}
	c.mu.Unlock()
	return umap
}

func (c *Channel) getMaxRoundTax() float64 {
	llog := c.l.With().Str("func", "getMaxRoundTax").Logger()

	if c.InspectionTax <= 0.0 { // use as a way to disable this functionality
		llog.Debug().
			Float64("InspectionTax", c.InspectionTax).
			Msg("Negative or zero InspectionTax, bailing out")
		return 0
	}
	lowestTotal := c.getLowestTotalInRound()
	if lowestTotal < 1 {
		llog.Debug().
			Int("lowestTotal", lowestTotal).
			Msg("Lowest total is below 1, bailing out")
		if err := c.postTaxFail("No tax today, as we have a participant with less than 1 points"); err != nil {
			llog.Error().Err(err).Send()
		}
		return 0
	}
	maxTax := (float64(lowestTotal) / 100.0) * c.InspectionTax
	if maxTax < 0.0 {
		llog.Debug().
			Float64("maxTax", maxTax).
			Msg("Calculated tax points is negative, bailing out")
		if err := c.postTaxFail(fmt.Sprintf("No tax today: Lowest total = %d, Percent = %f - which amounts to %f", lowestTotal, c.InspectionTax, maxTax)); err != nil {
			llog.Error().Err(err).Send()
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
	c.mu.Lock()
	c.InspectAlways = doInspect
	c.mu.Unlock()
}

func (c *Channel) getInspectAlways() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.InspectAlways
}

func (c *Channel) setTaxLoners(doTax bool) {
	c.mu.Lock()
	c.TaxLoners = doTax
	c.mu.Unlock()
}

func (c *Channel) getTaxLoners() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TaxLoners
}

func (c *Channel) shouldInspect() bool {
	llog := c.l.With().Str("func", "shouldInspect").Logger()
	// Having this check before the next will override TaxLoners
	if c.getInspectAlways() {
		llog.Debug().Msg("Configured to always run inspection")
		return true
	}
	// We could have something like this to only tax when more than 1 contestant
	if c.tmpNicks == nil || (!c.getTaxLoners() && len(c.tmpNicks) < 2) {
		llog.Debug().Msg("Configured to NOT tax loners")
		return false
	}

	wd := int(time.Now().Weekday())
	//nolint:gosec // sufficient
	rnd := rand.Intn(7) // 7 for number of days in week
	doInspect := wd == rnd

	llog.Debug().
		Int("weekday", wd).
		Int("rnd", rnd).
		Bool("inspect", doInspect).
		Msg("To inspect or not...")

	if !doInspect {
		if err := c.postTaxFail(fmt.Sprintf("No tax today :) Weekday = %d, random = %d", wd, rnd)); err != nil {
			llog.Error().Err(err).Send()
		}
	}

	return doInspect
}

// Return index in c.tmpNicks and how many points minus, if selected, otherwise -1 (or -2) and 0
func (c *Channel) randomInspect() (int, int) {
	llog := c.l.With().Str("func", "randomInspect").Logger()
	if !c.shouldInspect() {
		// unique "error" value indicating where this func bailed out
		return -2, 0
	}
	maxTax := c.getMaxRoundTax()
	if maxTax < 1 { // I don't think we've ever reached this section irl
		llog.Debug().
			Float64("maxTax", maxTax).
			Msg("Tax below 1, returning")
		if err := c.postTaxFail(fmt.Sprintf("No tax today. Calculated tax was: %f", maxTax)); err != nil {
			llog.Error().Err(err).Send()
		}
		return -1, 0
	}

	//nolint:gosec // sufficient
	return rand.Intn(len(c.tmpNicks)), rand.Intn(int(maxTax) + 1)
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
