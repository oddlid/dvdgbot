package l33t

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type BonusConfig struct {
	SubString    string // string to search for in timestamp
	Greeting     string // Message from bot to user upon bonus hit
	StepPoints   int    // points to multiply substring position with
	NoStepPoints int    // points to return for match when UseStep == false
	PrefixChar   rune   // the char required as only prefix for max bonus, e.g. '0'
	UseStep      bool   // if to multiply points for each position to the right in string
}

type BonusConfigs []BonusConfig

type BonusReturn struct {
	Match  string
	Msg    string
	Points int
}

type BonusReturns []BonusReturn

func (br BonusReturn) printBonus(w io.Writer) {
	fmt.Fprintf(w, "[%s=%d]: %s", br.Match, br.Points, br.Msg)
}

func (brs BonusReturns) totalBonus() int {
	total := 0
	for _, br := range brs {
		total += br.Points
	}
	return total
}

func (brs BonusReturns) printBonus(w io.Writer) {
	fmt.Fprintf(w, "+%d points bonus! : ", brs.totalBonus())
	for i, br := range brs {
		if i > 0 {
			fmt.Fprint(w, " + ")
		}
		br.printBonus(w)
	}
}

func hasHomogenicPrefix(ts string, prefix rune, matchPos int) bool {
	for i, r := range ts {
		if r != prefix {
			return false
		}
		if i >= matchPos-1 {
			break
		}
	}
	return true
}

// TODO: try to figure this out with math instead of string indexing
func (bc BonusConfig) calc(ts string) BonusReturn {
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
	matchPos := strings.Index(ts, bc.SubString)

	// There is no substring match, so we return empty value and don't bother with other checks
	if matchPos == -1 {
		return BonusReturn{}
	}

	br := BonusReturn{
		Points: bc.NoStepPoints,
		Match:  bc.SubString,
		Msg:    bc.Greeting,
	}
	// We have a match, but don't care about the substring position,
	// so we return points for any match without calculation
	if !bc.UseStep {
		return br
	}

	// We have a match, we DO care about position, but position is
	// 0, so we don't need to calculate, and can return StepPoints directly
	if matchPos == 0 {
		return br
	}

	// We have a match, we DO care about position, and position is above 0,
	// so now we need to calculate what to return

	// Position is not "purely prefixed" e.g. just zeros before the match
	if !hasHomogenicPrefix(ts, bc.PrefixChar, matchPos) {
		return br
	}

	// At this point, we know we have a match at position > 0, prefixed by only PrefixChar,
	// so we calculate bonus and return
	br.Points = (matchPos + 1) * bc.StepPoints
	return br
}

func (bcs BonusConfigs) calc(ts string) BonusReturns {
	brs := make(BonusReturns, 0)
	for _, bc := range bcs {
		br := bc.calc(ts)
		if br.Points > 0 {
			brs = append(brs, br)
		}
	}
	return brs
}

func (bcs BonusConfigs) greetForPoints(w io.Writer, points int) {
	for _, bc := range bcs {
		if val, err := strconv.Atoi(bc.SubString); err != nil {
			continue
		} else if val == points {
			fmt.Fprintf(w, " - %s", bc.Greeting)
			return
		}
	}
}
