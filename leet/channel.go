package leet

import (
	"sync"
)

type Channel struct {
	sync.RWMutex
	Users    map[string]*User `json:"users"`
	tmpNicks []string         // used for storing who participated in a specific round. Reset after calculation.
}

func (c *Channel) get(nick string) *User {
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
