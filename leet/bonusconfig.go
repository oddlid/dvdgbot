package leet

import (
	"strconv"
	"strings"
)

type BonusConfig struct {
	SubString    string // string to search for in timestamp
	PrefixChar   rune   // the char required as only prefix for max bonus, e.g. '0'
	UseStep      bool   // if to multiply points for each position to the right in string
	StepPoints   int    // points to multiply substring position with
	NoStepPoints int    // points to return for match when UseStep == false
	matchPos     int    // internal index for substring match position
}

type BonusConfigs []BonusConfig

func (bd BonusConfig) hasHomogenicPrefix(ts string) bool {
	for i, r := range ts {
		if r != bd.PrefixChar {
			return false
		}
		if i >= bd.matchPos-1 {
			break
		}
	}
	return true
}

func (bc BonusConfig) Calc(ts string) int {
	// We use the given hour and minute for point patterns.
	// The farther to the right the pattern occurs, the more points.
	// So, if hour = 13, minute = 37, we'd get something like this:
	// 13:37:13:37xxxxx = +(1 * STEP) points
	// 13:37:01:337xxxx = +(2 * STEP) points
	// 13:37:00:1337xxx = +(3 * STEP) points
	// 13:37:00:01337xx = +(4 * STEP) points
	// 13:37:00:001337x = +(5 * STEP) points
	// 13:37:00:0001337 = +(6 * STEP) points
	// ...


	// Search for substring match
	bc.matchPos = strings.Index(ts, bc.SubString)

	// There is no substring match, so we return 0 and don't bother with other checks
	if bc.matchPos == -1 {
		return 0
	}

	// We have a match, but don't care about the substring position,
	//so we return points for any match without calculation
	if !bc.UseStep {
		return bc.NoStepPoints
	}

	// We have a match, we DO care about position, but position is
	// 0, so we don't need to calculate, and can return StepPoints directly
	if bc.matchPos == 0 {
		return bc.StepPoints
	}

	// We have a match, we DO care about position, and position is above 0,
	// so now we need to calculate what to return

	// Position is not "purely prefixed" e.g. just zeros before the match
	if !bc.hasHomogenicPrefix(ts) {
		return bc.StepPoints
	}

	// At this point, we know we have a match at position > 0, prefixed by only PrefixChar,
	// so we calculate bonus and return
	return (bc.matchPos + 1) * bc.StepPoints
}

func (bcs *BonusConfigs) Add(bc BonusConfig) {
	*bcs = append(*bcs, bc)
}

func (bcs BonusConfigs) Calc(ts string) int {
	total := 0
	for _, bc := range bcs {
		total += bc.Calc(ts)
	}
	return total
}

func (bcs BonusConfigs) HasValue(val int) bool {
	for _, bc := range bcs {
		ival, err := strconv.Atoi(bc.SubString)
		if err != nil {
			continue
		}
		if ival == val {
			return true
		}
	}
	return false
}
