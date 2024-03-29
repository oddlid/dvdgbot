package leet

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/rs/zerolog"
)

const (
	testChannel = "#blackhole"
)

var (
	boolVar bool
	intVar  int
	strVar  string
	tcVar   TimeCode
	userVar *User
)

func getData() *ScoreData {
	// in case this is run via gotest.sh, then we'd have a copy of real data here to use
	if _scoreData != nil && !_scoreData.isEmpty() {
		return _scoreData
	}

	if _scoreData == nil {
		_scoreData = newScoreData()
	}

	// Fill with some test data if empty
	// fmt.Println("Creating Oddlid")
	// o := _scoreData.get(TST_CHAN).get("Oddlid")
	// fmt.Println("Creating Tord")
	// t := _scoreData.get(TST_CHAN).get("Tord")
	// fmt.Println("Creating Snelhest")
	// s := _scoreData.get(TST_CHAN).get("Snelhest")
	// fmt.Println("Creating bAAAArd")
	// b := _scoreData.get(TST_CHAN).get("bAAAArd")

	// o.addScore(10)
	// t.addScore(8)
	// s.addScore(6)
	// b.addScore(4)

	nicks := []string{
		"Oddlid",
		"Snelhest",
		"Tord",
		"Tormod",
		"ZeGerman",
		"bAAAArd",
		"bitslap",
	}

	for idx, nick := range nicks {
		points := idx * 2
		_scoreData.get(testChannel).get(nick).score(points, time.Now())
	}

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

func TestInspection(_ *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.InspectionTax = 50.14 // % of total points for the user with the least points in the current round
	for k := range c.Users {
		c.addNickForRound(k) // adds to c.tmpNicks
	}

	for i := 0; i < 100; i++ {
		nickIdx, tax := c.randomInspect()
		if nickIdx < 0 {
			continue
		}
		nick := c.tmpNicks[nickIdx]
		c.get(nick).
			l.Info().
			Int("iteration", i).
			Int("tax", tax).
			Msg("Selected for inspection")
	}
}

func TestInspectLoner(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.InspectionTax = 100.0 // % of total points for the user with the least points in the current round
	nick := "Oddlid"
	c.clearNicksForRound() // forgetting this made tests fail when running with ./...
	c.addNickForRound(nick)

	c.setInspectAlways(true)
	if !c.shouldInspect() {
		t.Errorf("Set to always inspect, but shouldInspect() returned false anyhow")
	}

	c.setInspectAlways(false)
	c.setTaxLoners(false)
	for i := 0; i < 10; i++ {
		if c.shouldInspect() {
			t.Errorf("Set to not inspect loners, but did so anyway. len(c.tmpNicks) = %d", len(c.tmpNicks))
		}
	}

	c.setTaxLoners(true)
	llog := c.get(nick).l
	for i := 0; i < 10; i++ {
		llog.Info().
			Bool("shouldInspect", c.shouldInspect()).
			Msg("Inspect?")
	}
}

func TestTaxFail(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.InspectionTax = 50.14 // % of total points for the user with the least points in the current round
	c.PostTaxFail = true
	for k := range c.Users {
		c.addNickForRound(k) // adds to c.tmpNicks
	}

	if c.Name != testChannel {
		t.Errorf("Expected channel name %q, got %q", testChannel, c.Name)
	}

	c.randomInspect()
}

func TestStats(_ *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()

	u := sd.get(testChannel).get("Oddlid")
	u.lock()
	u.setLastEntry(time.Now())
	u.setScore(getTargetScore())

	// add a BonusConfig that matches the current time
	_bonusConfigs.add(
		BonusConfig{
			SubString:    fmt.Sprintf("%d", getTargetScore()),
			Greeting:     "The final goal has been reached!",
			PrefixChar:   48,
			UseStep:      false,
			StepPoints:   0,
			NoStepPoints: 5,
		},
	)

	fmt.Printf("%s", sd.stats(testChannel))
}

func TestGetTargetScore(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	defScore := getTargetScore()

	t.Logf("Target score, from ENV (or default): %d\n", defScore)

	// Since the result is cached in _targetScore, we need to reset that before altering
	// _hour and _minute if we want changes to apply

	// Verify caching, that modifying these without modifying _targetScore, does nothing
	_hour = 12
	_minute = 34

	if getTargetScore() != defScore {
		t.Errorf("Expected %d, but got %d", defScore, getTargetScore())
	}

	// Now make sure value is recalculated after resetting _targetScore
	_targetScore = 0
	_hour = 12
	_minute = 0o4 // have a prefixing 0 to make sure it's included in the result

	exp := 1204
	score := getTargetScore()
	if score != exp {
		t.Errorf("Expected %d, but got %d", exp, score)
	} else {
		t.Logf("Target score after manual intervention is: %d", score)
	}
}

func TestWinner(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)

	nick := "Oddlid"
	user := c.get(nick)
	user.setScore(getTargetScore())
	user.setLastEntry(time.Now())
	user.lock()

	locked := user.isLocked()
	if locked {
		rank := c.getWinnerRank(nick)
		tx := timexDiff(sd.BotStart, user.getLastEntry())
		fmt.Printf(
			"%s: You're locked, as you're #%d, reaching %d points @ %s after %s :)\n",
			nick,
			rank+1,
			getTargetScore(),
			getLongDate(user.getLastEntry()),
			tx.String(),
		)
	} else {
		t.Errorf("User %s expected to be locked", nick)
	}
}

func TestUserSort(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)

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

	osmap := c.getOverShooters(targetScore) // this should include all above but nicks[0]

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
		if pu == nil {
			pu = v
		} else if v.LastEntry.Before(pu.LastEntry) {
			t.Errorf("%s is after %s", v.LastEntry, pu.LastEntry)
		}
		fmt.Printf(
			"%-10s: %d @ %s\n",
			v.Nick,
			v.Points,
			v.LastEntry.Format(tFmt),
		)
	}
}

func TestOverShooters(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.OvershootTax = 5
	limit := getTargetScore()

	nicks := c.nickList()

	// Helper for seeing who hit the spot right on
	// When we get to the real code, we should when we check overshooters, check if the one at 0 points
	// isLocked(), and if so, just ignore, otherwise add to list of winners.
	markWinner := func(points int) (int, string) {
		if points == 0 {
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

	umap := c.getOverShooters(limit)

	fmt.Printf("\nLimit     : %d\n\n", limit)

	for idx, nick := range nicks {
		u := umap[nick]
		expPoints := limit + idx*2
		if u.getScore() != expPoints {
			t.Errorf("Expected %q to have %d points, but got %d", nick, expPoints, u.getScore())
		}
		points, mark := markWinner(expPoints - limit)
		fmt.Printf("%-10s: %d (+ %d points)%s\n", nick, u.getScore(), points, mark)
	}

	fmt.Printf("\n")
	c.punishOverShooters(limit, umap)

	for idx, nick := range nicks {
		u := umap[nick]
		upoints := limit + idx*2
		tax := c.getOverShootTaxFor(limit, upoints)
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

	umap = c.getOverShooters(limit)

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

// Here we'll try to simulate a round where users pass the finish line
func TestRaceToFinish(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.OvershootTax = 15
	limit := getTargetScore()
	nicks := c.nickList()
	var sb strings.Builder

	t.Logf("Target points: %d", limit)
	t.Log("")

	for idx, nick := range nicks {
		startingPoints := limit + (idx - 2)
		timeAdjVal := (idx - 2) * 17
		t.Logf("Nick: %q, starts with %d points, time adjustment: %d", nick, startingPoints, timeAdjVal)
		u := c.get(nick)
		u.setScore(startingPoints)
		et := time.Now().Add(time.Duration(timeAdjVal) * time.Second)
		success, msg := sd.tryScore(c, u, et)
		if success {
			t.Logf("Bot reply: %q", msg)
		}
	}

	// At this point, we should have some users below the limit, maybe some just at the spot, and some over.
	// Depending on the exact time the test is run, some will get minus points for before and after, and often,
	// some will get bonus points, no matter the placement.
	// IRL, this is where sd.calcAndPost() would run, but we'll simulate that here instad of calling it, so we
	// can get that func correctly implemented later.

	// This is the first part of calcAndPost(), which calcs points and syncs them to the users.
	// We skip the message generation right now.
	scoreMap := c.getScoresForRound()
	c.mergeScoresForRound(scoreMap)

	t.Log("\nUser points after round calculation:")

	for _, nick := range nicks {
		t.Logf("%-10s: %d", nick, c.get(nick).getScore())
	}

	// Punish overshooters
	// IRL, we'd might like to do this in more steps in order to post back messages to those
	// with bad luck...
	umap := c.punishOverShooters(limit, c.getOverShooters(limit))

	t.Log("\nUser points after overshoot taxation:")

	for _, nick := range nicks {
		t.Logf("%-10s: %d", nick, c.get(nick).getScore())
	}

	// do tax
	idx, tax := c.randomInspect() // most times we get -1 here and skip the rest
	if idx > -1 {
		nick := c.tmpNicks[idx]
		user := c.get(nick)
		if tax > 0 {
			user.addScore(-tax)
			fmt.Fprintf(
				&sb,
				"%s was randomly selected for taxation and lost %d points (now: %d points)",
				nick,
				tax,
				user.getScore(),
			)
		} else {
			fmt.Fprintf(&sb, "%s was randomly selected for taxation, but got off with a slap on the wrist ;)", nick)
		}
		// msgChan(channel, sb.String())
		t.Log(sb.String())
	}

	ws := umap.filterByPointsEQ(limit)
	// Add winners to channel
	for _, user := range ws {
		user.lock()
	}

	ws.sortByLastEntryAsc()

	t.Log("\nWinners:")
	for idx, user := range ws {
		t.Logf("%-10s: %d #%d [%s]", user.Nick, user.getScore(), idx+1, getShortTime(user.getLastEntry()))
	}

	// clean up
	c.clearNicksForRound()
}

func TestCalcScore(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	c.OvershootTax = 15
	limit := getTargetScore()
	nicks := c.nickList()

	// add a BonusConfig that matches the current time
	_bonusConfigs.add(
		BonusConfig{
			SubString:    fmt.Sprintf("%d", limit),
			Greeting:     "The final goal has been reached!",
			PrefixChar:   48,
			UseStep:      false,
			StepPoints:   0,
			NoStepPoints: 5,
		},
	)

	t.Logf("Target points: %d", limit)
	t.Log("")

	for idx, nick := range nicks {
		startingPoints := limit + (idx - 3)
		timeAdjVal := (idx - 2) * 17
		t.Logf("Nick: %q, starts with %d points, time adjustment: %d", nick, startingPoints, timeAdjVal)
		u := c.get(nick)
		u.setScore(startingPoints)
		et := time.Now().Add(time.Duration(timeAdjVal) * time.Second)
		success, msg := sd.tryScore(c, u, et)
		if success {
			t.Logf("Bot reply: %q", msg)
		}
	}

	t.Logf("\n%s", sd.calcScore(c))
}

func TestSetBestEntry(_ *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)

	user := c.get("Oddlid")

	oldTime := time.Now().Add(-(2 * time.Minute))
	newTime := oldTime.Add(-(1 * time.Second))

	user.setLastEntry(oldTime)

	user.setBestEntry(newTime)

	for i := 1; i < 30; i++ {
		newTime = newTime.Add(time.Duration(i) * time.Second)
		user.setBestEntry(newTime)
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

	var exp, got int

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
		// got = bcs.calc(stamps[i])
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
	// got = bcs.calc(ts)
	brs := bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 28
	ts = stamps[10]
	// got = bcs.calc(ts)
	brs = bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 68
	ts = stamps[12]
	// got = bcs.calc(ts)
	brs = bcs.calc(ts)
	got = brs.TotalBonus()
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}
}

func TestGetCronTime(t *testing.T) {
	h, m := getCronTime(13, 37, -2*time.Minute)
	if h != 13 || m != 35 {
		t.Errorf("Expected 13 35, but got: %d %d", h, m)
	}
	h, m = getCronTime(12, 0o0, -3*time.Minute)
	if h != 11 || m != 57 {
		t.Errorf("Expected 11 57, but got: %d %d", h, m)
	}
}

// func TestNtpCheck(t *testing.T) {
// 	zerolog.SetGlobalLevel(zerolog.DebugLevel)
// 	h, m := getCronTime(_hour, _minute, 1*time.Minute)
// 	success := scheduleNtpCheck(h, m, "0.se.pool.ntp.org")
// 	if success {
// 		t.Log("Sleeping for 70 seconds...")
// 		time.Sleep(70 * time.Second)
// 		t.Logf("NTP offset after scheduled update: %+v", _ntpOffset)
// 	}
// }

func TestLastTsInCurrentRound(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	u := c.get("Oddlid")

	now := time.Now()
	u.setLastEntry(now.AddDate(0, 0, -1)) // set to 1 day before now

	// t.Logf("User last entry: %+v", u.getLastEntry())

	if u.lastTSInCurrentRound(now) {
		t.Error("1 day after lastEntry should not count as being in current round")
	}

	u.setLastEntry(now)

	if !u.lastTSInCurrentRound(now) {
		t.Error("Equal times should count as in current round")
	}
	if !u.lastTSInCurrentRound(now.Add(1 * time.Minute)) {
		t.Error("1 minute after lastEntry should count as in current round")
	}
	if !u.lastTSInCurrentRound(now.Add(2 * time.Minute)) {
		t.Error("2 minutes after lastEntry should count as in current round")
	}
	if !u.lastTSInCurrentRound(now.Add(3 * time.Minute)) {
		t.Error("3 minutes after lastEntry should count as in current round")
	}
	if u.lastTSInCurrentRound(now.Add(4 * time.Minute)) {
		t.Error("4 minutes after lastEntry should NOT count as in current round")
	}
}

// We should bench both calling the method repeatedly and also implementing
// the same locally so we have cached values, so we can see how much waste
// it is to call that method to get only one rank.
// After running these, it showed the wasteful method getWinnerRank to be
// over 20 times slower than caching the filtered and sorted slice, and
// then just call getIndex for the given nick.
func BenchmarkGetWinnerRank(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	limit := getTargetScore()

	for _, u := range c.Users {
		u.setScore(limit)
		u.setLastEntry(time.Now())
		u.lock()
	}

	rank := 0
	// b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for nick := range c.Users {
			rank = c.getWinnerRank(nick)
		}
	}
	intVar = rank
}

// Locally cached version for comparison
func BenchmarkGetWinnerRankCached(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)
	limit := getTargetScore()

	for _, u := range c.Users {
		u.setScore(limit)
		u.setLastEntry(time.Now())
		u.lock()
	}

	// Now we have done the filtering in advance and have the results locally,
	// so we can see how much is wasted on doing this repeatedly when calling c.getWinnerRank
	us := c.Users.filterByLocked(true)
	us.sortByLastEntryAsc()

	rank := 0
	// b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for nick := range c.Users {
			rank = us.getIndex(nick)
		}
	}
	intVar = rank
}

/*
2021-02-23:
I was thinking that I should maybe redesign this whole module, as there's constantly
shitloads of lookups of channel and user throughout the code.
But, since most of them are for existing channel/user, and seeing as getting an existing
map entry is so much faster than creating new ones, I don't think it is really important.
At least not in the performance scope of this game.
But, just as an exercise in making better code designs, it would be good.
We should then get both the channel and user object in leet(), as that is already done,
then pass those pointers around everywhere they're used. That should save a lot of lookups.
*/

// 24.3 ns/op
func BenchmarkGetExistingUser(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)

	nick := "Oddlid"

	var user *User
	for i := 0; i < b.N; i++ {
		user = c.get(nick)
	}
	userVar = user
}

// 820 ns/op
func BenchmarkGetNonExistingUser(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	sd := getData()
	c := sd.get(testChannel)

	nick := "Oddlid"
	delete(c.Users, nick)

	var user *User
	for i := 0; i < b.N; i++ {
		user = c.get(nick)
		delete(c.Users, nick)
	}
	userVar = user
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
			// got = bcs.calc(ts)
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
		ts   time.Time
		nick string
	}{
		{nick: "Odd_01", ts: tm("2019-04-07T13:37:00.000001337Z")},
		{nick: "Odd_02", ts: tm("2019-04-07T13:37:00.000013370Z")},
		{nick: "Odd_03", ts: tm("2019-04-07T13:37:00.000133700Z")},
	}
	c := sd.get(channel)
	var bres bool
	var sres string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, n := range nicks {
			bres, sres = sd.tryScore(c, c.get(n.nick), n.ts)
		}
		// c.MergeScoresForRound(c.GetScoresForRound())
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

	// var result string
	for i := 0; i < b.N; i++ {
		cmd.User.Nick = fmt.Sprintf("Nick_%d", i)
		_, err := leet(cmd)
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
		// b.Log(msg)
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
		if strings.Contains(msg, "bonus") {
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
		if strings.Contains(msg, "[1337") {
			b.Log(msg)
		}
	}
}
