package leet

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/oddlid/dvdgbot/util"
)

// Constants used for module settings, unless corresponding env vars are given
const (
	defaultHour      = 13                               // Override with env var LEETBOT_HOUR
	defaultMinute    = 37                               // Override with env var LEETBOT_MINUTE
	scoreFile        = "/tmp/leetbot_scores.json"       // Override with env var LEETBOT_SCOREFILE
	bonusConfigsFile = "/tmp/leetbot_bonusconfigs.json" // Override with env var LEETBOT_BONUSCONFIGFILE
	plugin           = "LeetBot"                        // Just used for log output
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
	_log             = log.With().Str("plugin", plugin).Logger()
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
	_log.Debug().
		Str("func", "msgChan").
		Str("channel", channel).
		Str("message", msg).
		Send()
	if _bot == nil {
		return fmt.Errorf("parentBot is nil")
	}
	if channel == "" {
		return fmt.Errorf("no channel name given")
	}
	if msg == "" {
		return fmt.Errorf("refusing to send empty message")
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
		return tcBefore
	} else if th > _hour {
		return tcAfter
	}
	// now we know we're within the correct hour
	tm := t.Minute()
	if tm < _minute-1 {
		return tcBefore
	} else if tm > _minute+1 {
		return tcAfter
	} else if tm == _minute-1 {
		return tcEarly
	} else if tm == _minute+1 {
		return tcLate
	}
	return tcOnTime
}

func withinTimeFrame(t time.Time) (bool, TimeCode) {
	tf := timeFrame(t)
	if tf == tcEarly || tf == tcOnTime || tf == tcLate {
		return true, tf
	}
	return false, tf
}

// getScoreForEntry returns "0, TF_ONTIME" if you should get points,
// and "-1, TF_(BEFORE|EARLY|LATE|AFTER)" if you miss and should not have points
func getScoreForEntry(t time.Time) (int, TimeCode) {
	var points int
	tf := timeFrame(t)

	if tf == tcEarly || tf == tcLate {
		points = -1
	} else {
		points = 0 // will be set later if on time
	}

	return points, tf
}

func checkArgs(cmd *bot.Cmd) (bool, string) {
	llog := _log.With().Str("func", "checkArgs").Logger()
	alen := len(cmd.Args)
	if alen == 1 && cmd.Args[0] == "stats" {
		if _scoreData.calcInProgress {
			return false, "Stats are calculating. Try again in a couple of minutes."
		}
		return false, _scoreData.stats(cmd.Channel)
	} else if alen == 1 && cmd.Args[0] == "reload" {
		// TODO: Handle load errors and give feedback for BC as well

		if err := _bonusConfigs.loadFile(_bonusConfigFile); err != nil {
			llog.Error().
				Err(err).
				Msg("Error loading Bonus Configs from file")
		}
		if !_scoreData.saveInProgress {
			_, err := _scoreData.loadFile(_scoreFile)
			if err != nil {
				llog.Error().Err(err).Send()
				return false, err.Error()
			}
			return false, "Score data reloaded from file"
		}
		return false, "A scheduled save is in progress. Will not reload right now."
	} else if alen >= 1 {
		return false, fmt.Sprintf("Unrecognized argument: %q. Usage: !1337 [stats|reload]", cmd.Args[0])
	}
	return true, ""
}

func leet(cmd *bot.Cmd) (string, error) {
	t := time.Now() // save time as early as possible

	proceed, msg := checkArgs(cmd)
	if !proceed {
		return strings.TrimRight(msg, "\n"), nil
	}

	// Adjust time for NTP offset, if set
	if _ntpOffset != 0 {
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
		tx := timexDiff(_scoreData.BotStart, u.getLastEntry())
		return fmt.Sprintf(
			"%s: You're locked, as you're #%d, reaching %d points @ %s after %s :)",
			u.Nick,
			c.getWinnerRank(u.Nick),
			u.getScore(),
			getLongDate(u.getLastEntry()),
			tx.String(),
		), nil
	}

	// is the user spamming?
	if u.hasTried() {
		return fmt.Sprintf("%s: Stop spamming!", u.Nick), nil
	}

	// this call also saves the users last entry time, which is important later
	success, msg := _scoreData.tryScore(c, u, t)

	// at this point, data might have changed, and should be saved
	var delay int
	switch tf {
	case tcEarly:
		delay = 3
	case tcOnTime:
		delay = 2
	case tcLate:
		delay = 1
	default:
		delay = 0
	}

	if success && !_scoreData.saveInProgress {
		_scoreData.scheduleSave(_scoreFile, time.Duration(delay+1)*time.Minute)
	}

	if !_scoreData.calcInProgress && c.hasPendingScores() {
		_scoreData.scheduleCalcScore(c, time.Duration(delay)*time.Minute)
	}

	if success {
		return strings.TrimRight(msg, "\n"), nil
	}

	// bogus
	return "", fmt.Errorf("%s: Reached beyond logic", plugin)
}

// getTargetScore should be used only _after_ pickupEnv, as it will modify
// vars _hour/_minute if not set
func getTargetScore() int {
	// return cached result if it exists
	if _targetScore != 0 {
		return _targetScore
	}
	if _hour == 0 {
		_hour = defaultHour
	}
	if _minute == 0 {
		_minute = defaultMinute
	}
	intVal, err := strconv.Atoi(fmt.Sprintf("%d%02d", _hour, _minute))
	if err != nil {
		_log.Error().
			Err(err).
			Msg("Conversion error")
		return -1
	}
	// cache the result for later calls
	_targetScore = intVal
	return _targetScore
}

// Helper to make sure we get the right time, even when adjusting across time borders
func getCronTime(hour, minute int, adjust time.Duration) (int, int) {
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
	return when.Hour(), when.Minute()
}

func scheduleNtpCheck(hour, minute int, server string) bool {
	// Since this func is called from init(), we need to use log.(Info|Error) here,
	// as we haven't yet reached the point where log.DebugLevel is set.

	llog := _log.With().
		Str("func", "scheduleNtpCheck").
		Str("server", server).
		Int("hour", hour).
		Int("minute", minute).
		Logger()

	if server == "" {
		llog.Info().Msg("Empty server, skipping scheduling")
		return false
	}

	if hour < 0 || hour > 23 {
		llog.Error().Msg("Hour must be between 0 and 23")
		return false
	}

	if minute < 0 || minute > 59 {
		llog.Error().Msg("Minute must be between 0 and 59")
		return false
	}

	llog.Info().Msg("Setting up cronjob")
	if _cron == nil {
		_cron = cron.New()
	}

	cronSpec := fmt.Sprintf("%d %d * * *", minute, hour)
	llog.Info().
		Str("cronSpec", cronSpec).
		Msg("Setting CRON SPEC")

	id, err := _cron.AddFunc(
		cronSpec,
		func() {
			llog.Info().Msg("Running NTP query...")
			offset, err := getNtpOffset(server)
			if err != nil {
				_ntpOffset = 0 // reset, so we don't use offset that might be way off since last sync
				llog.Error().Err(err).Send()
				return
			}
			llog.Info().
				Dur("ntpOffset", offset).
				Msg("Updating NTP offset")
			_ntpOffset = offset
			// notify all channels
			msg := fmt.Sprintf("NTP offset from %q: %+v", server, _ntpOffset)
			for channel := range _scoreData.Channels {
				if err := msgChan(channel, msg); err != nil {
					llog.Error().Err(err).Msgf("Failed to send message to channel %q", channel)
				}
			}
		},
	)
	if err != nil {
		llog.Error().Err(err).Send()
		return false
	}
	llog.Info().
		Int("entryID", int(id)).
		Msg("Cronjob successfully setup, starting cron")

	_cron.Start()

	return true
}

func pickupEnv() {
	_hour = util.EnvDefInt("LEETBOT_HOUR", defaultHour)
	_minute = util.EnvDefInt("LEETBOT_MINUTE", defaultMinute)
	_scoreFile = util.EnvDefStr("LEETBOT_SCOREFILE", scoreFile)
	_bonusConfigFile = util.EnvDefStr("LEETBOT_BONUSCONFIGFILE", bonusConfigsFile)
	_ntpServer = util.EnvDefStr("LEETBOT_NTP_SERVER", "") // we want empty as default if not specified here
}

func init() {
	pickupEnv()

	var err error
	llog := _log.With().Str("func", "init").Logger()

	_scoreData, err = newScoreData().loadFile(_scoreFile)
	if err != nil {
		llog.Error().
			Err(err).
			Msg("Error loading scoredata from file")
	}

	if err = _bonusConfigs.loadFile(_bonusConfigFile); err != nil {
		llog.Error().
			Err(err).
			Msg("Error loading Bonus Configs from file")
	}

	// This is hard to test, as I'd normally set _hour and _minute to the same time as when running tests,
	// so if testing, comment out this temporarily, or just ignore it, as it would be 24 hours - 2 minutes
	// until it triggers.
	// I had an idea that one could set a variable via -X to the linker when compiling, and then just do this
	// if the variable is set, which should prevent this from being run during tests and so on, but for now,
	// I can live with it being like this.
	// But during testing one may also just not set env LEETBOT_NTP_SERVER, and if so, this section is ignored.
	if _ntpServer != "" {
		llog.Info().
			Str("ntpServer", _ntpServer).
			Msg("NTP server configured, scheduling NTP checks...")
		h, m := getCronTime(_hour, _minute, -2*time.Minute)
		ok := scheduleNtpCheck(h, m, _ntpServer)
		if ok {
			llog.Info().Msg("NTP check scheduled")
		} else {
			llog.Error().Msg("Error scheduling NTP check")
		}
	} else {
		llog.Info().Msg("No NTP server set")
	}

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats|reload]",
		leet,
	)
}
