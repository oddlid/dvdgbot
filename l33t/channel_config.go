package l33t

import (
	"math/rand"
	"time"
)

type ChannelData struct {
	Users         UserDataMap
	InspectionTax float64 `json:"inspection_tax"`
	OvershootTax  int     `json:"overshoot_tax"`
	InspectAlways bool    `json:"inspect_always"`
	TaxLoners     bool    `json:"tax_loners"`
	PostTaxFail   bool    `json:"post_tax_fail"`
}

type ChannelDataMap map[string]*ChannelData

type InspectionDecision struct {
	Inspect bool
	Weekday time.Weekday
	Random  time.Weekday
}

// getChannel returns a Channel for the data associated with the name if it exists,
// otherwise it creates and inserts an empty data entry for the given name and
// returns a new Channel for that
func (ccm ChannelDataMap) getChannel(name string) *Channel {
	if ccm == nil {
		return nil
	}
	channelData, ok := ccm[name]
	if ok {
		return &Channel{
			name: name,
			data: channelData,
		}
	}

	ccm[name] = &ChannelData{
		Users: make(UserDataMap),
	}
	return ccm.getChannel(name)
}

func (cc *ChannelData) empty() bool {
	return cc == nil || (len(cc.Users) == 0 && cc.InspectionTax == 0.0 && cc.OvershootTax == 0 && !cc.InspectAlways && !cc.TaxLoners && !cc.PostTaxFail)
}

func (cc ChannelData) shouldInspect(t time.Time, numContestants int) InspectionDecision {
	if cc.InspectAlways {
		return InspectionDecision{
			Inspect: true,
			Weekday: -1,
			Random:  -1,
		}
	}
	if !cc.TaxLoners && numContestants < 2 {
		return InspectionDecision{
			Inspect: false,
			Weekday: -1,
			Random:  -1,
		}
	}

	ret := InspectionDecision{
		Weekday: t.Weekday(),
		Random:  time.Weekday(rand.Intn(int(time.Saturday) + 1)), //nolint:gosec // sufficient
	}
	ret.Inspect = ret.Weekday == ret.Random
	return ret
}
