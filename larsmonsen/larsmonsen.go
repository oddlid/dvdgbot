package larsmonsen

import (
	"regexp"

	"github.com/go-chat-bot/bot"
	"github.com/oddlid/dvdgbot/quoteshuffle"
	"github.com/oddlid/dvdgbot/util"
	"github.com/rs/zerolog/log"
)

const (
	PLUGIN     = "LarsMonsen"
	FACTS_FILE = "/tmp/larsmonsenfacts.json"
	PATTERN    = "(?i)\\b(lars|monsen)\\b"
	CMD_NAME   = "larsmonsen"
)

var (
	qd   *quoteshuffle.QuoteData
	re   = regexp.MustCompile(PATTERN)
	_log = log.With().Str("plugin", PLUGIN).Logger()
)

func larsmonsen(command *bot.PassiveCmd) (string, error) {
	if re.MatchString(command.Raw) {
		return qd.Shuffle(), nil
	}
	return "", nil
}

func init() {
	var err error
	qd, err = quoteshuffle.New(util.EnvDefStr("LARSMONSENFACTS_FILE", FACTS_FILE))
	if err != nil {
		_log.Error().
			Err(err).
			Msg("Error loading facts file")
		return
	}
	bot.RegisterPassiveCommand(CMD_NAME, larsmonsen)
}
