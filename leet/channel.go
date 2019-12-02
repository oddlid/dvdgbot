package leet

import (
	"sync"
	"time"
)

type Channel struct {
	sync.RWMutex
	Users    map[string]*User `json:"users"`
	tmpNicks []string         // used for storing who participated in a specific round. Reset after calculation.
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
