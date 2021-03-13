package leet

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Constants used for module settings, unless corresponding env vars are given
const (
	DEF_HOUR          = 13                               // Override with env var LEETBOT_HOUR
	DEF_MINUTE        = 37                               // Override with env var LEETBOT_MINUTE
	SCORE_FILE        = "/tmp/leetbot_scores.json"       // Override with env var LEETBOT_SCOREFILE
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
	_targetScore     int
	_scoreFile       string
	_bonusConfigFile string
	_scoreData       *ScoreData
	_bot             *bot.Bot
	_bonusConfigs    BonusConfigs
	_log             = logrus.WithField("plugin", PLUGIN) // *logrus.Entry
	_ntpServer       string
	_ntpOffset       time.Duration
	_cron            *cron.Cron
)

// SetParentBot sets the internal global reference to an instance of "github.com/go-chat-bot/bot".
// This reference is used to send messages outside of the registered callback (leet).
// The bot including this module must call this before using this module.
func SetParentBot(b *bot.Bot) {
	_bot = b
}

func msgChan(channel, msg string) error {
	if nil == _bot {
		str := "ParentBot is nil"
		_log.WithFields(logrus.Fields{
			"func":    "msgChan",
			"channel": channel,
			"message": msg,
		}).Error(str)
		return fmt.Errorf(str)
	}
	_bot.SendMessage(
		bot.OutgoingMessage{
			Target:      channel,
			Message:     msg,
			Sender:      &bot.User{},
			ProtoParams: nil,
		},
	)
	return nil
}

func getPadStrFmt(alignAt int, format string) string {
	return fmt.Sprintf("%s%d%s", "%-", alignAt, "s "+format)
}

func writePad(w io.Writer, align int, str string) {
	strfmt := getPadStrFmt(align, "")
	fmt.Fprintf(w, strfmt, str)
}

func inStrSlice(slc []string, val string) (int, bool) {
	for idx, entry := range slc {
		if val == entry {
			return idx, true
		}
	}
	return -1, false
}

func getLongDate(t time.Time) string {
	// use 0's to get subsecond value padded,
	// use 9's to get trailing 0's removed.
	// I don't know yet why, but when running tests on my mac,
	// I always get the last 3 digits as 0 when using padding,
	// although they're never 0 when calling t.Nanoseconds()
	// other places in the code.
	// TODO: Try to get full precision here
	return t.Format("2006-01-02 15:04:05.000000000")
}

func getShortTime(t time.Time) string {
	return t.Format("15:04:05.000000000")
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

// getScoreForEntry returns "0, TF_ONTIME" if you should get points,
// and "-1, TF_(BEFORE|EARLY|LATE|AFTER)" if you miss and should not have points
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
	llog := _log.WithField("func", "checkArgs")
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
			llog.WithError(err).Error("Error loading Bonus Configs from file")
		}
		if !_scoreData.saveInProgress {
			_, err = _scoreData.loadFile(_scoreFile)
			if nil != err {
				llog.Error(err)
				msg = err.Error()
			} else {
				msg = "Score data reloaded from file"
			}
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

	// Adjust time for NTP offset, if set
	if 0 != _ntpOffset {
		// Tempting to add a log statement here, but seeing as slow as that is, we don't
		// want to lose time to that in this func
		t = t.Add(_ntpOffset)
	}

	// don't give a fuck outside accepted time frame
	inTimeFrame, tf := withinTimeFrame(t)
	if !inTimeFrame {
		return "", nil
	}

	// has the user already reached the target point sum and should not contend?
	c := _scoreData.get(cmd.Channel)
	u := c.get(cmd.User.Nick)
	if u.isLocked() {
		return fmt.Sprintf(
			"%s: You're locked, as you're #%d, reaching %d points @ %s after %s :)",
			u.Nick,
			c.getWinnerRank(u.Nick),
			u.getScore(),
			getLongDate(u.getLastEntry()),
			timexString(timexDiff(_scoreData.BotStart, u.getLastEntry())),
		), nil
	}

	// is the user spamming?
	if u.hasTried() {
		return fmt.Sprintf("%s: Stop spamming!", u.Nick), nil
	}

	// this call also saves the users last entry time, which is important later
	success, msg := _scoreData.tryScore(c, u, t)

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

	if !_scoreData.calcInProgress && c.hasPendingScores() {
		_scoreData.scheduleCalcScore(c, delayMinutes)
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

// getTargetScore should be used only _after_ pickupEnv, as it will modify
// vars _hour/_minute if not set
func getTargetScore() int {
	// return cached result if it exists
	if 0 != _targetScore {
		return _targetScore
	}
	if 0 == _hour {
		_hour = DEF_HOUR
	}
	if 0 == _minute {
		_minute = DEF_MINUTE
	}
	intVal, err := strconv.Atoi(fmt.Sprintf("%d%02d", _hour, _minute))
	if nil != err {
		_log.WithError(err).Error("Conversion error")
		return -1
	}
	// cache the result for later calls
	_targetScore = intVal
	return _targetScore
}

// Helper to make sure we get the right time, even when adjusting across
// time borders
func getCronTime(hour, minute int, adjust time.Duration) (h, m int) {
	now := time.Now()
	then := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hour,
		minute,
		now.Second(),
		now.Nanosecond(),
		now.Location(),
	)
	when := then.Add(adjust)
	h = when.Hour()
	m = when.Minute()
	return
}

func scheduleNtpCheck(hour, minute int, server string) bool {
	// Since this func is called from init(), we need to use log.(Info|Error) here,
	// as we haven't yet reached the point where log.DebugLevel is set.

	llog := _log.WithFields(logrus.Fields{
		"func":   "scheduleNtpCheck",
		"hour":   hour,
		"minute": minute,
		"server": server,
	})

	if "" == server {
		llog.Info("Empty server, skipping scheduling")
		return false
	}

	if hour < 0 || hour > 23 {
		llog.Error("Hour must be between 0 and 23")
		return false
	}

	if minute < 0 || minute > 59 {
		llog.Error("Minute must be between 0 and 59")
		return false
	}

	llog.Info("Setting up cronjob")
	if nil == _cron {
		_cron = cron.New()
	}

	cronSpec := fmt.Sprintf("%d %d * * *", minute, hour)
	llog.WithField("cronSpec", cronSpec).Info("Setting CRON SPEC")

	id, err := _cron.AddFunc(
		cronSpec,
		func() {
			llog.Info("Running NTP query...")
			offset, err := getNtpOffset(server)
			if nil != err {
				llog.Error(err)
				return
			}
			llog.WithField("ntpOffset", offset).Info("Updating NTP offset")
			_ntpOffset = offset
			// notify all channels
			msg := fmt.Sprintf("NTP offset from %q: %+v", server, _ntpOffset)
			for channel, _ := range _scoreData.Channels {
				msgChan(channel, msg)
			}
		},
	)
	if nil != err {
		llog.Error(err)
		return false
	}
	llog.WithField("entryID", id).Info("Cronjob successfully setup")

	llog.Info("Starting cron")
	_cron.Start()

	return true
}

func pickupEnv() {
	_hour = envDefInt("LEETBOT_HOUR", DEF_HOUR)
	_minute = envDefInt("LEETBOT_MINUTE", DEF_MINUTE)
	_scoreFile = envDefStr("LEETBOT_SCOREFILE", SCORE_FILE)
	_bonusConfigFile = envDefStr("LEETBOT_BONUSCONFIGFILE", BONUSCONFIGS_FILE)
	_ntpServer = envDefStr("LEETBOT_NTP_SERVER", "") // we want empty as default if not specified here
}

func init() {
	pickupEnv()

	var err error
	llog := _log.WithField("func", "init")

	_scoreData, err = newScoreData().loadFile(_scoreFile)
	if err != nil {
		llog.WithError(err).Error("Error loading scoredata from file")
	}

	err = _bonusConfigs.loadFile(_bonusConfigFile)
	if err != nil {
		llog.WithError(err).Error("Error loading Bonus Configs from file")
	}

	// This is hard to test, as I'd normally set _hour and _minute to the same time as when running tests,
	// so if testing, comment out this temporarily, or just ignore it, as it would be 24 hours - 2 minutes
	// until it triggers.
	// I had an idea that one could set a variable via -X to the linker when compiling, and then just do this
	// if the variable is set, which should prevent this from being run during tests and so on, but for now,
	// I can live with it being like this.
	if "" != _ntpServer {
		llog.WithField("ntpServer", _ntpServer).Info("NTP server configured, scheduling NTP checks...")
		h, m := getCronTime(_hour, _minute, -2*time.Minute)
		ok := scheduleNtpCheck(h, m, _ntpServer)
		if ok {
			llog.Info("NTP check scheduled")
		} else {
			llog.Error("Error scheduling NTP check")
		}
	} else {
		llog.Info("No NTP server set")
	}

	// Init rand for using in tax calculation
	rand.Seed(time.Now().UnixNano())

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
}
