package leet

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/sirupsen/logrus"
)

// Constants used for module settings, unless corresponding env vars are given
const (
	DEF_HOUR          = 13                               // Override with env var LEETBOT_HOUR
	DEF_MINUTE        = 37                               // Override with env var LEETBOT_MINUTE
	SCORE_FILE        = "/tmp/leetbot_scores.json"       // Override with env var LEETBOT_MINUTE
	BONUSCONFIGS_FILE = "/tmp/leetbot_bonusconfigs.json" // Override with env var LEETBOT_BONUSCONFIGFILE
	PLUGIN            = "LeetBot"                        // Just used for log output
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
	_log             = logrus.WithField("plugin", PLUGIN) // *logrus.Entry
)

// SetParentBot sets the internal global reference to an instance of "github.com/go-chat-bot/bot".
// This reference is used to send messages outside of the registered callback (leet).
// The bot including this module must call this before using this module.
func SetParentBot(b *bot.Bot) {
	_bot = b
}

func msgChan(channel, msg string) error {
	if nil == _bot {
		msg := "ParentBot is nil"
		_log.Error(msg)
		return fmt.Errorf(msg)
	}
	_bot.SendMessage(
		bot.OutgoingMessage{
			channel,
			msg,
			&bot.User{},
			nil,
		},
	)
	return nil
}

func getPadStrFmt(alignAt int, format string) string {
	return fmt.Sprintf("%s%d%s", "%-", alignAt, "s "+format)
}

func longestEntryLen(s []string) int {
	maxlen := 0
	for i := range s {
		nlen := len(s[i])
		if nlen > maxlen {
			maxlen = nlen
		}
	}
	return maxlen
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
	if TF_EARLY == tf || TF_ONTIME == tf || TF_LATE == tf {
		return true, tf
	}
	return false, tf
}

func getScoreForEntry(t time.Time) (int, TimeCode) {
	var points int
	tf := timeFrame(t)

	if TF_EARLY == tf || TF_LATE == tf {
		points = -1
	} else {
		points = 0 // will be set later if on time
	}

	return points, tf
}

func checkArgs(cmd *bot.Cmd) (proceed bool, msg string) {
	alen := len(cmd.Args)
	if alen == 1 && "stats" == cmd.Args[0] {
		if _scoreData.calcInProgress {
			msg = "Stats are calculating. Try again in a couple of minutes."
			return
		} else {
			msg = _scoreData.stats(cmd.Channel)
			return
		}
	} else if alen == 1 && "reload" == cmd.Args[0] {
		// TODO: Handle load errors and give feedback for BC as well
		err := _bonusConfigs.loadFile(_bonusConfigFile)
		if err != nil {
			_log.WithError(err).Error("Error loading Bonus Configs from file")
		}
		if !_scoreData.saveInProgress {
			_scoreData.loadFile(_scoreFile)
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
		return strings.TrimRight(msg, "\n"), nil
	}

	// don't give a fuck outside accepted time frame
	inTimeFrame, tf := withinTimeFrame(t)
	if !inTimeFrame {
		return "", nil
	}

	// is the user spamming?
	if _scoreData.didTry(cmd.Channel, cmd.User.Nick) {
		return fmt.Sprintf("%s: Stop spamming!", cmd.User.Nick), nil
	}

	success, msg := _scoreData.tryScore(cmd.Channel, cmd.User.Nick, t)

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
		_scoreData.scheduleSave(_scoreFile, delayMinutes+1)
	}

	if !_scoreData.calcInProgress && _scoreData.get(cmd.Channel).hasPendingScores() {
		_scoreData.scheduleCalcScore(cmd.Channel, delayMinutes)
	}

	if success {
		return strings.TrimRight(msg, "\n"), nil
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

	_scoreData, err = newScoreData().loadFile(_scoreFile)
	if err != nil {
		_log.WithError(err).Error("Error loading scoredata from file")
	}

	err = _bonusConfigs.loadFile(_bonusConfigFile)
	if err != nil {
		_log.WithError(err).Error("Error loading Bonus Configs from file")
	}

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
	// Init rand for picking out the sore loser of the day
	rand.Seed(time.Now().UnixNano())
}
