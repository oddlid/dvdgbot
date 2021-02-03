package leet

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chat-bot/bot"
	log "github.com/sirupsen/logrus"
)

const (
	TST_CHAN = "#blackhole"
)

var (
	boolVar bool
	intVar  int
	strVar  string
	tcVar   TimeCode
)

func getData() *ScoreData {
	// in case this is run via gotest.sh, then we'd have a copy or real data here to use
	if nil != _scoreData && !_scoreData.isEmpty(){
		return _scoreData
	}

	if nil == _scoreData {
		_scoreData = newScoreData()
	}

	// Fill with some test data if empty
	fmt.Println("Creating Oddlid")
	o := _scoreData.get(TST_CHAN).get("Oddlid")
	fmt.Println("Creating Tord")
	t := _scoreData.get(TST_CHAN).get("Tord")
	fmt.Println("Creating Snelhest")
	s := _scoreData.get(TST_CHAN).get("Snelhest")
	fmt.Println("Creating bAAAArd")
	b := _scoreData.get(TST_CHAN).get("bAAAArd")

	o.score(10)
	t.score(8)
	s.score(6)
	b.score(4)

	return _scoreData
}

func TestSave(t *testing.T) {
	const fname string = "/tmp/leetdata.json"
	file, err := os.Create(fname)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	sd := getData()
	n, err := sd.save(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Saved %d bytes to %q", n, fname)
}

//func TestInspection(t *testing.T) {
//	sd := getData()
//	c := sd.get(TST_CHAN)
//	c.InspectionTax = 5
//	for k, _ := range c.Users {
//		c.addNickForRound(k)
//	}
//	t.Logf("Had  | withdrawn | now  | nick")
//	for k, v := range c.Users {
//		_, sub := c.randomInspect()
//		total := v.Points - sub
//		t.Logf("%04d | %04d      | %04d | %s", v.Points, sub, total, k)
//		if total < 0 {
//			t.Log("Subtracted beyond 0")
//			t.FailNow()
//		}
//	}
//}

func TestInspection(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	sd := getData()
	c := sd.get(TST_CHAN)
	c.InspectionTax = 50.14 // % of total points for the user with the least points in the current round
	for k, _ := range c.Users {
		c.addNickForRound(k) // adds to c.tmpNicks
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		nickIdx, tax := c.randomInspect()
		if nickIdx < 0 {
			continue
		}
		nick := c.tmpNicks[nickIdx]
		c.get(nick).log().WithFields(log.Fields{
			"iteration": i,
			"tax":       tax,
		}).Info("Selected for inspection")
	}
}

func TestInspectLoner(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	sd := getData()
	c := sd.get(TST_CHAN)
	c.InspectionTax = 100.0 // % of total points for the user with the least points in the current round
	rand.Seed(time.Now().UnixNano())
	nick := "Oddlid"
	c.addNickForRound(nick)

	c.InspectAlways = true
	if !c.shouldInspect() {
		t.Errorf("Set to always inspect, but shouldInspect() returned false anyhow")
	}

	c.InspectAlways = false
	c.TaxLoners = false
	for i := 0; i < 10; i++ {
		if c.shouldInspect() {
			t.Errorf("Set to not inspect loners, but did so anyway")
		}
	}

	c.TaxLoners = true
	llog := c.get(nick).log()
	for i := 0; i < 10; i++ {
		llog.WithFields(log.Fields{
			"shouldInspect": c.shouldInspect(),
		}).Info("Inspect?")
	}
}

func TestTaxFail(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	c := sd.get(TST_CHAN)
	c.InspectionTax = 50.14 // % of total points for the user with the least points in the current round
	c.PostTaxFail = true
	for k, _ := range c.Users {
		c.addNickForRound(k) // adds to c.tmpNicks
	}

	cname, err := c.name()
	if nil != err {
		t.Error(err)
	}
	if cname != TST_CHAN {
		t.Errorf("Expected channel name %q, got %q", TST_CHAN, cname)
	}
	c.randomInspect()
}

func TestStats(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	//c := sd.get(TST_CHAN)

	fmt.Printf("%s", sd.stats(TST_CHAN))
}

func TestWinner(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	c := sd.get(TST_CHAN)

	// actual botstart: 2018-03-07T17:27:15.435563057+02:00
	//bStart, tErr := time.Parse(time.RFC3339, "2018-03-07T17:27:15+02:00")
	//if nil == tErr {
	//	sd.BotStart = bStart
	//}


	nick1 := "Oddlid"
	//nick2 := "Tord"
	odd := c.get(nick1)
	//tord := c.get(nick2)
	odd.setScore(getTargetScore())
	odd.setLastEntry(time.Now())
	//tord.setScore(getTargetScore())
	sd.addWinner(TST_CHAN, nick1)
	//sd.addWinner(TST_CHAN, nick2, time.Now())

	locked, rating := sd.isLocked(TST_CHAN, nick1)
	if locked {
		fmt.Printf(
			"%s: You're locked, as you're #%d, reaching %d points @ %s after %s :)\n",
			nick1,
			rating.Rank + 1,
			getTargetScore(),
			rating.ReachedAt.Format("2006-01-02 15:04:05.999999999"),
			timexString(timexDiff(sd.BotStart, rating.ReachedAt)),
		)
	} else {
		t.Errorf("User %s expected to be locked", nick1)
	}
}

func TestUserSort(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	c := sd.get(TST_CHAN)

	nicks := c.nickList()
	targetScore := getTargetScore()
	entryTime := time.Now()
	tFmt := "15:04:05.999999999"

	u0 := c.get(nicks[0])
	u0.setScore(targetScore - 1)
	u0.setLastEntry(entryTime)

	u1 := c.get(nicks[1])
	u1.setScore(targetScore)
	u1.setLastEntry(entryTime.Add(100 * time.Millisecond))

	u2 := c.get(nicks[2])
	u2.setScore(targetScore)
	u2.setLastEntry(entryTime.Add(200 * time.Millisecond))

	u3 := c.get(nicks[3])
	u3.setScore(targetScore)
	u3.setLastEntry(entryTime.Add(300 * time.Millisecond))

	u4 := c.get(nicks[4])
	u4.setScore(targetScore + 1)
	u4.setLastEntry(entryTime.Add(400 * time.Millisecond))

	osmap := c.getOverShooters() // this should include all above but nicks[0]

	fmt.Printf("\nTarget score: %d\n", targetScore)

	fmt.Printf("\nUndershooter:\n")
	fmt.Printf(
		"%-10s: %d @ %s\n",
		u0.Nick,
		u0.Points,
		u0.LastEntry.Format(tFmt),
	)

	fmt.Printf("\nOvershooters:\n")
	for _, v := range osmap {
		fmt.Printf(
			"%-10s: %d @ %s\n",
			v.Nick,
			v.Points,
			v.LastEntry.Format(tFmt),
		)
	}

	ws := osmap.filterByPointsEQ(targetScore)
	fmt.Printf("\nWinners, unsorted:\n")
	for _, v := range ws {
		fmt.Printf(
			"%-10s: %d @ %s\n",
			v.Nick,
			v.Points,
			v.LastEntry.Format(tFmt),
		)
	}

	ws.sortByLastEntryAsc()
	var pu *User
	fmt.Printf("\nWinners, sorted:\n")
	for _, v := range ws {
		if nil == pu {
			pu = v
		} else {
			if v.LastEntry.Before(pu.LastEntry) {
				t.Errorf("%s is after %s", v.LastEntry, pu.LastEntry)
			}
		}
		fmt.Printf(
			"%-10s: %d @ %s\n",
			v.Nick,
			v.Points,
			v.LastEntry.Format(tFmt),
		)
	}
}

func TestRemoveNickFromRound(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	c := sd.get(TST_CHAN)

	//nicks := []string{
	//	"Oddlid",
	//	"Tord",
	//	"Snelhest",
	//	"bAAAArd",
	//}
	nicks := c.nickList() // swap this for the above if not loading local list
	for _, n := range nicks {
		c.addNickForRound(n)
	}
	origLen := len(c.tmpNicks)
	nickToRemove := nicks[1]

	c.removeNickFromRound(nickToRemove)

	newLen := len(c.tmpNicks)

	// check length
	if newLen != origLen - 1 {
		t.Errorf("Expected length to be %d, but got %d", origLen - 1, newLen)
	}

	// check that it does not contain removed nick
	for _, n := range c.tmpNicks {
		if n == nickToRemove {
			t.Errorf("%q should not be in c.tmpNicks anymore", nickToRemove)
		}
	}

	// check order preserved
	if c.tmpNicks[0] != nicks[0] {
		t.Errorf("Nick at index %d should be %q", 0, nicks[0])
	}
	if c.tmpNicks[1] != nicks[2] {
		t.Errorf("Nick at index %d should be %q", 1, nicks[2])
	}
	if c.tmpNicks[2] != nicks[3] {
		t.Errorf("Nick at index %d should be %q", 2, nicks[3])
	}

}

func TestOverShooters(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
	sd := getData()
	c := sd.get(TST_CHAN)
	c.OvershootTax = 5
	limit := getTargetScore()

	//nicks := []string{
	//	"Oddlid",
	//	"Tord",
	//	"Snelhest",
	//	"bAAAArd",
	//}
	nicks := c.nickList()

	// Helper for seeing who hit the spot right on
	// When we get to the real code, we should when we check overshooters, check if the one at 0 points
	// isLocked(), and if so, just ignore, otherwise add to list of winners.
	markWinner := func(points int) (int, string) {
		if 0 == points {
			return points, " - Winner!"
		}
		return points, ""
	}

	// First, test when everyone got at or over the limit

	for idx, nick := range nicks {
		u := c.get(nick)
		u.setLastEntry(time.Now())
		u.setScore(limit + idx*2)
	}

	umap := c.getOverShooters()

	fmt.Printf("\nLimit     : %d\n\n", limit)

	for idx, nick := range nicks {
		u := umap[nick]
		expPoints := limit + idx*2
		if u.getScore() != expPoints {
			t.Errorf("Expected %q to have %d points, but got %d", nick, expPoints, u.getScore())
		}
		points, mark := markWinner(expPoints-limit)
		fmt.Printf("%-10s: %d (+ %d points)%s\n", nick, u.getScore(), points, mark)
	}

	fmt.Printf("\n")
	c.punishOverShooters(umap)

	for idx, nick := range nicks {
		u := umap[nick]
		upoints := limit + idx*2
		tax := c.getOverShootTaxFor(upoints)
		expPoints := upoints - tax
		if u.getScore() != expPoints {
			t.Errorf("Expected %q to have %d points, but got %d", nick, expPoints, u.getScore())
		}
		fmt.Printf("%-10s: %d (- %d points)\n", nick, u.getScore(), tax)
	}

	// Second, test when some are below and some at or over

	for idx, nick := range nicks {
		u := c.get(nick)
		u.setLastEntry(time.Now())
		u.setScore((limit - 1) + idx*2) // first nick should be below the limit
	}

	umap = c.getOverShooters()

	fmt.Printf("\nLimit     : %d\n\n", limit)

	for idx, nick := range nicks {
		u, found := umap[nick]
		if !found {
			score := c.get(nick).getScore()
			fmt.Printf("%-10s: %d (%d points below limit)\n", nick, score, limit-score)
			continue
		}
		expPoints := (limit - 1) + idx*2
		if u.getScore() != expPoints {
			t.Errorf("Expected %q to have %d points, but got %d", nick, expPoints, u.getScore())
		}
		fmt.Printf("%-10s: %d (+ %d points)\n", nick, u.getScore(), expPoints-limit)
	}
}

func TestBonusConfigCalc(t *testing.T) {
	var bcs BonusConfigs
	stamps := []string{
		"00000000000",
		"13370000000",
		"01337000000",
		"00133700000",
		"00013370000",
		"00001337000",
		"00000133700",
		"00000013370",
		"00000001337",
		"01000001337",
		"00006661337",
		"00006660000",
		"00001337666",
	}

	exp := 0
	got := 0

	bc1 := BonusConfig{
		SubString:    "1337",
		PrefixChar:   '0',
		UseStep:      true,
		StepPoints:   10,
		NoStepPoints: 0,
	}
	bc2 := BonusConfig{
		SubString:    "666",
		PrefixChar:   '0',
		UseStep:      false,
		StepPoints:   0,
		NoStepPoints: 18,
	}

	bcs.add(bc1)

	for i := 0; i <= 8; i++ {
		exp = i * bc1.StepPoints
		//got = bcs.calc(stamps[i])
		brs := bcs.calc(stamps[i])
		got = brs.TotalBonus()
		if got != exp {
			t.Errorf("Expected %d, got %d from substring %s", exp, got, stamps[i])
		}
		t.Logf("%s gives %d points bonus", stamps[i], got)
	}

	bcs.add(bc2)

	exp = 18
	ts := stamps[11]
	//got = bcs.calc(ts)
	brs := bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 28
	ts = stamps[10]
	//got = bcs.calc(ts)
	brs = bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 68
	ts = stamps[12]
	//got = bcs.calc(ts)
	brs = bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

}

// This BM shows that almost all execution time in bonus() goes to
// 2 lines of string formatting...
func BenchmarkStrFmt(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var str1 string
	var str2 string

	for i := 0; i < b.N; i++ {
		str1 = fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
		str2 = fmt.Sprintf("%02d%02d", 13, 37)
	}
	strVar = str1
	strVar = str2
}

func BenchmarkStrIndex(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	tstr := fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
	sstr := "1337"
	var ires int

	for i := 0; i < b.N; i++ {
		ires = strings.Index(tstr, sstr)
	}
	intVar = ires
}

//func BenchmarkPrefixedBy(b *testing.B) {
//	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
//	tstr := fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
//	sstr := "1337"
//	idx := strings.Index(tstr, sstr)
//	var bres bool
//
//	for i := 0; i < b.N; i++ {
//		bres = pb(tstr, '0', idx)
//	}
//	boolVar = bres
//}

//func BenchmarkBonus(b *testing.B) {
//	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
//	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:13.370000000Z")
//	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:01.337000000Z")
//	//bp := 0
//	var ires int
//	for i := 0; i < b.N; i++ {
//		//bp = bonus(ts)
//		ires = bonus(ts)
//	}
//	intVar = ires
//	//b.Logf("bonus: %d", bp)
//}

func BenchmarkBonusConfigCalc(b *testing.B) {
	var bcs BonusConfigs
	stamps := []string{
		"00000000000",
		"13370000000",
		"01337000000",
		"00133700000",
		"00013370000",
		"00001337000",
		"00000133700",
		"00000013370",
		"00000001337",
		"01000001337",
		"00006661337",
		"00006660000",
		"00001337666",
	}

	bc1 := BonusConfig{
		SubString:    "1337",
		PrefixChar:   '0',
		UseStep:      true,
		StepPoints:   10,
		NoStepPoints: 0,
	}
	bc2 := BonusConfig{
		SubString:    "666",
		PrefixChar:   '0',
		UseStep:      false,
		StepPoints:   0,
		NoStepPoints: 18,
	}

	bcs.add(bc1)
	bcs.add(bc2)

	got := 0

	for _, ts := range stamps {
		for i := 0; i < b.N; i++ {
			//got = bcs.calc(ts)
			brs := bcs.calc(ts)
			got = brs.TotalBonus()
		}
		intVar = got
	}
}

func BenchmarkTryScore(b *testing.B) {
	tm := func(tstr string) time.Time {
		t, _ := time.Parse(time.RFC3339Nano, tstr)
		return t
	}
	sd := newScoreData()
	channel := "#channel"
	nicks := []struct {
		nick string
		ts   time.Time
	}{
		{"Odd_01", tm("2019-04-07T13:37:00.000001337Z")},
		{"Odd_02", tm("2019-04-07T13:37:00.000013370Z")},
		{"Odd_03", tm("2019-04-07T13:37:00.000133700Z")},
	}
	c := sd.get(channel)
	var bres bool
	var sres string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, n := range nicks {
			bres, sres = sd.tryScore(channel, n.nick, n.ts)
		}
		//c.MergeScoresForRound(c.GetScoresForRound())
		c.clearNicksForRound()
	}
	boolVar = bres
	strVar = sres
}

func BenchmarkWithinTimeFrame(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var bres bool
	var tcres TimeCode
	for i := 0; i < b.N; i++ {
		bres, tcres = withinTimeFrame(ts)
	}
	boolVar = bres
	tcVar = tcres
}

func BenchmarkTimeFrame(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var result TimeCode
	for i := 0; i < b.N; i++ {
		result = timeFrame(ts)
	}
	tcVar = result
}

func BenchmarkDidTry(b *testing.B) {
	sd := newScoreData()
	nick := "Odd"
	channel := "#channel"
	var result bool

	for i := 0; i < b.N; i++ {
		result = sd.didTry(channel, nick)
	}
	boolVar = result
}

func BenchmarkLeet(b *testing.B) {
	// I'd like to see if I can mock the whole process, and see how tight posts could get
	// I don't quite know what's required to fill in to the Cmd struct, so we'll see...
	cmd := &bot.Cmd{
		Raw:     "",
		Channel: "#channel",
		ChannelData: &bot.ChannelData{
			Protocol:  "",
			Server:    "",
			Channel:   "#channel",
			HumanName: "",
			IsPrivate: false,
		},
		User: &bot.User{
			ID:       "",
			Nick:     "",
			RealName: "",
			IsBot:    false,
		},
		Message: "!1337",
		MessageData: &bot.Message{
			Text:     "!1337",
			IsAction: false,
			ProtoMsg: nil,
		},
		Command: "",
		RawArgs: "",
		Args:    nil,
	}

	//var result string
	for i := 0; i < b.N; i++ {
		cmd.User.Nick = fmt.Sprintf("Nick_%d", i)
		msg, err := leet(cmd)
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
		b.Log(msg)
	}
}

func BenchmarkHitBonus(b *testing.B) {
	cmd := &bot.Cmd{
		Raw:     "",
		Channel: "#channel",
		ChannelData: &bot.ChannelData{
			Protocol:  "",
			Server:    "",
			Channel:   "#channel",
			HumanName: "",
			IsPrivate: false,
		},
		User: &bot.User{
			ID:       "",
			Nick:     "",
			RealName: "",
			IsBot:    false,
		},
		Message: "!1337",
		MessageData: &bot.Message{
			Text:     "!1337",
			IsAction: false,
			ProtoMsg: nil,
		},
		Command: "",
		RawArgs: "",
		Args:    nil,
	}
	for i := 0; i < b.N; i++ {
		cmd.User.Nick = fmt.Sprintf("Nick_%d", i)
		msg, err := leet(cmd)
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
		if strings.Index(msg, "bonus") > -1 {
			b.Log(msg)
		}
	}
}

func BenchmarkHit1337(b *testing.B) {
	cmd := &bot.Cmd{
		Raw:     "",
		Channel: "#channel",
		ChannelData: &bot.ChannelData{
			Protocol:  "",
			Server:    "",
			Channel:   "#channel",
			HumanName: "",
			IsPrivate: false,
		},
		User: &bot.User{
			ID:       "",
			Nick:     "",
			RealName: "",
			IsBot:    false,
		},
		Message: "!1337",
		MessageData: &bot.Message{
			Text:     "!1337",
			IsAction: false,
			ProtoMsg: nil,
		},
		Command: "",
		RawArgs: "",
		Args:    nil,
	}
	for i := 0; i < b.N; i++ {
		cmd.User.Nick = fmt.Sprintf("Nick_%d", i)
		msg, err := leet(cmd)
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
		if strings.Index(msg, "[1337") > -1 {
			b.Log(msg)
		}
	}
}
