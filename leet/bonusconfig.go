package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type BonusConfig struct {
	SubString    string // string to search for in timestamp
	Greeting     string // Message from bot to user upon bonus hit
	StepPoints   int    // points to multiply substring position with
	NoStepPoints int    // points to return for match when UseStep == false
	matchPos     int    // internal index for substring match position
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

func (br BonusReturn) String() string {
	return fmt.Sprintf("[%s=%d]: %s", br.Match, br.Points, br.Msg)
}

func (brs BonusReturns) TotalBonus() int {
	total := 0
	for _, br := range brs {
		total += br.Points
	}
	return total
}

func (brs BonusReturns) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "+%d points bonus! : ", brs.TotalBonus())
	for i, br := range brs {
		if i > 0 {
			sb.WriteString(" + ")
		}
		sb.WriteString(br.String())
	}
	return sb.String()
}

func (bc BonusConfig) hasHomogenicPrefix(ts string) bool {
	for i, r := range ts {
		if r != bc.PrefixChar {
			return false
		}
		if i >= bc.matchPos-1 {
			break
		}
	}
	return true
}

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
	bc.matchPos = strings.Index(ts, bc.SubString)

	// There is no substring match, so we return empty value and don't bother with other checks
	if bc.matchPos == -1 {
		return BonusReturn{}
	}

	// We have a match, but don't care about the substring position,
	// so we return points for any match without calculation
	if !bc.UseStep {
		return BonusReturn{
			Points: bc.NoStepPoints,
			Match:  bc.SubString,
			Msg:    bc.Greeting,
		}
	}

	// We have a match, we DO care about position, but position is
	// 0, so we don't need to calculate, and can return StepPoints directly
	if bc.matchPos == 0 {
		return BonusReturn{
			Points: bc.StepPoints,
			Match:  bc.SubString,
			Msg:    bc.Greeting,
		}
	}

	// We have a match, we DO care about position, and position is above 0,
	// so now we need to calculate what to return

	// Position is not "purely prefixed" e.g. just zeros before the match
	if !bc.hasHomogenicPrefix(ts) {
		return BonusReturn{
			Points: bc.StepPoints,
			Match:  bc.SubString,
			Msg:    bc.Greeting,
		}
	}

	// At this point, we know we have a match at position > 0, prefixed by only PrefixChar,
	// so we calculate bonus and return
	return BonusReturn{
		Points: (bc.matchPos + 1) * bc.StepPoints,
		Match:  bc.SubString,
		Msg:    bc.Greeting,
	}
}

func (bcs *BonusConfigs) add(bc BonusConfig) {
	*bcs = append(*bcs, bc)
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

func (bcs BonusConfigs) hasValue(val int) (bool, BonusConfig) {
	for _, bc := range bcs {
		ival, err := strconv.Atoi(bc.SubString)
		if err != nil {
			continue
		}
		if ival == val {
			return true, bc
		}
	}
	return false, BonusConfig{}
}

func (bcs *BonusConfigs) load(r io.Reader) error {
	jb, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, bcs)
}

func (bcs *BonusConfigs) loadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return bcs.load(file)
}

func (bcs BonusConfigs) save(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(bcs, "", "\t")
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (bcs BonusConfigs) saveFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = bcs.save(file)
	if err != nil {
		return err
	}
	return nil
}
