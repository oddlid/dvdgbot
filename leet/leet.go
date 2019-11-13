package leet

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

const (
	DEF_HOUR          = 13
	DEF_MINUTE        = 37
	BONUS_STEP        = 10
	SCORE_FILE        = "/tmp/leetbot_scores.json"
	BONUSCONFIGS_FILE = "/tmp/leetbot_bonusconfigs.json"
	PLUGIN            = "LeetBot"
)

type TimeCode int

// Constants for signalling how much off timeframe
const (
	TF_BEFORE TimeCode = iota // more than a minute before
	TF_EARLY                  // less than a minute before
	TF_ONTIME                 // within correct minute
	TF_LATE                   // less than a minute late
	TF_AFTER                  // more than a minute late
)

var (
	_hour         int = DEF_HOUR
	_minute       int = DEF_MINUTE
	_scoreData    *ScoreData
	_bot          *bot.Bot
	_bonusConfigs BonusConfigs
)


func SetParentBot(b *bot.Bot) {
	_bot = b
}

func timeFrame(t time.Time) TimeCode {
	th := t.Hour()
	if th < _hour {
		return TF_BEFORE
	} else if th > _hour {
		return TF_AFTER
	}
	// now we know we're within the correct hour
	tm := t.Minute()
	if tm < _minute-1 {
		return TF_BEFORE
	} else if tm > _minute+1 {
		return TF_AFTER
	} else if tm == _minute-1 {
		return TF_EARLY
	} else if tm == _minute+1 {
		return TF_LATE
	}
	return TF_ONTIME
}

func withinTimeFrame(t time.Time) (bool, TimeCode) {
	tf := timeFrame(t)
	if tf == TF_EARLY || tf == TF_ONTIME || tf == TF_LATE {
		return true, tf
	}
	return false, tf
}

func leet(cmd *bot.Cmd) (string, error) {
	t := time.Now() // save time as early as possible

	// handle arguments
	alen := len(cmd.Args)
	if alen == 1 && "stats" == cmd.Args[0] {
		if _scoreData.calcInProgress {
			return "Stats are calculating. Try again in a couple of minutes.", nil
		} else {
			return _scoreData.Stats(cmd.Channel), nil
		}
	} else if alen == 1 && "reload" == cmd.Args[0] {
		// TODO: Handle load errors and give feedback for BC as well
		err := _bonusConfigs.LoadFile(BONUSCONFIGS_FILE)
		if err != nil {
			log.Error(err)
		}
		if !_scoreData.saveInProgress {
			_scoreData.LoadFile(SCORE_FILE)
			return "Score data reloaded from file", nil
		} else {
			return "A scheduled save is in progress. Will not reload right now.", nil
		}
	} else if alen >= 1 {
		return fmt.Sprintf("Unrecognized argument: %q. Usage: !1337 [stats|reload]", cmd.Args[0]), nil
	}

	// don't give a fuck outside accepted time frame
	inTimeFrame, tf := withinTimeFrame(t)
	if !inTimeFrame {
		return "", nil
	}

	// is the user spamming?
	if _scoreData.DidTry(cmd.Channel, cmd.User.Nick) {
		return fmt.Sprintf("%s: Stop spamming!", cmd.User.Nick), nil
	}

	success, msg := _scoreData.TryScore(cmd.Channel, cmd.User.Nick, t)

	// at this point, data might have changed, and should be saved
	var delayMinutes time.Duration
	if TF_EARLY == tf {
		delayMinutes = 3
	} else if TF_ONTIME == tf {
		delayMinutes = 2
	} else if TF_LATE == tf {
		delayMinutes = 1
	}

	if success && !_scoreData.saveInProgress {
		_scoreData.ScheduleSave(SCORE_FILE, delayMinutes+1)
	}

	if !_scoreData.calcInProgress && _scoreData.Get(cmd.Channel).HasPendingScores() {
		_scoreData.ScheduleCalcScore(cmd.Channel, delayMinutes)
	}

	if success {
		return msg, nil
	}

	// bogus
	return "", fmt.Errorf("%s: Reached beyond logic...", PLUGIN)
}

func pickupEnv() {
	h := os.Getenv("LEETBOT_HOUR")
	m := os.Getenv("LEETBOT_MINUTE")

	var err error
	if h != "" {
		_hour, err = strconv.Atoi(h)
		if err != nil {
			_hour = DEF_HOUR
		}
	}
	if m != "" {
		_minute, err = strconv.Atoi(m)
		if err != nil {
			_minute = DEF_MINUTE
		}
	}
}

func init() {
	_scoreData = NewScoreData().LoadFile(SCORE_FILE)
	pickupEnv() // for minute/hour. IMPORTANT: this has to come before bonusconfigs, as they use these values to generate strings

//	_bonusConfigs.Add(
//		BonusConfig{
//			SubString:    fmt.Sprintf("%02d%02d", _hour, _minute), // '1337' when used as intended
//			PrefixChar:   '0',
//			UseStep:      true,
//			StepPoints:   10,
//			NoStepPoints: 0,
//		},
//	)
//	_bonusConfigs.Add(
//		BonusConfig{
//			SubString:    "666", // because, of course...
//			PrefixChar:   '0',   // not used, as UseStep is false
//			UseStep:      false, //
//			StepPoints:   0,     // not used
//			NoStepPoints: 18,    // 18 points, because 6+6+6 = 18
//		},
//	)

	err := _bonusConfigs.LoadFile(BONUSCONFIGS_FILE)
	if err != nil {
		log.Error(err)
	}

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
}
