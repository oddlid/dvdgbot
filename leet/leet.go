package leet

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-chat-bot/bot"
	log "github.com/sirupsen/logrus"
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
	_hour            int
	_minute          int
	_scoreFile       string
	_bonusConfigFile string
	_scoreData       *ScoreData
	_bot             *bot.Bot
	_bonusConfigs    BonusConfigs
	_log             = log.WithField("plugin", PLUGIN)
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

func checkArgs(cmd *bot.Cmd) (proceed bool, msg string) {
	alen := len(cmd.Args)
	if alen == 1 && "stats" == cmd.Args[0] {
		if _scoreData.calcInProgress {
			msg = "Stats are calculating. Try again in a couple of minutes."
			return
		} else {
			msg = _scoreData.Stats(cmd.Channel)
			return
		}
	} else if alen == 1 && "reload" == cmd.Args[0] {
		// TODO: Handle load errors and give feedback for BC as well
		err := _bonusConfigs.LoadFile(_bonusConfigFile)
		if err != nil {
			_log.WithError(err).Error("Error lading Bonus Configs from file")
		}
		if !_scoreData.saveInProgress {
			_scoreData.LoadFile(_scoreFile)
			msg = "Score data reloaded from file"
			return
		} else {
			msg = "A scheduled save is in progress. Will not reload right now."
			return
		}
	} else if alen >= 1 {
		msg = fmt.Sprintf("Unrecognized argument: %q. Usage: !1337 [stats|reload]", cmd.Args[0])
		return
	}
	proceed = true
	return
}

func leet(cmd *bot.Cmd) (string, error) {
	t := time.Now() // save time as early as possible

	proceed, msg := checkArgs(cmd)
	if !proceed {
		return msg, nil
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
		_scoreData.ScheduleSave(_scoreFile, delayMinutes+1)
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

func envDefStr(key, fallback string) string {
	val, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return val // might still be empty, if set, but empty in ENV
}

func envDefInt(key string, fallback int) int {
	val, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		_log.WithError(err).Error("Conversion error")
		return fallback
	}
	return intVal
}

func pickupEnv() {
	_hour = envDefInt("LEETBOT_HOUR", DEF_HOUR)
	_minute = envDefInt("LEETBOT_MINUTE", DEF_MINUTE)
	_scoreFile = envDefStr("LEETBOT_SCOREFILE", SCORE_FILE)
	_bonusConfigFile = envDefStr("LEETBOT_BONUSCONFIGFILE", BONUSCONFIGS_FILE)
}

func init() {
	pickupEnv()

	var err error

	_scoreData, err = NewScoreData().LoadFile(_scoreFile)
	if err != nil {
		_log.WithError(err).Error("Error loading scoredata from file")
	}

	err = _bonusConfigs.LoadFile(_bonusConfigFile)
	if err != nil {
		_log.WithError(err).Error("Error loading Bonus Configs from file")
	}

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
}
