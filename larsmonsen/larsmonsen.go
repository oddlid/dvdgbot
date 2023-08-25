package larsmonsen

import (
	"regexp"

	"github.com/go-chat-bot/bot"
	"github.com/oddlid/dvdgbot/quoteshuffle"
	"github.com/oddlid/dvdgbot/util"
	"github.com/rs/zerolog/log"
)

const (
	plugin    = "LarsMonsen"
	factsFile = "/tmp/larsmonsenfacts.json"
	pattern   = "(?i)\\b(lars|monsen)\\b"
	cmdName   = "larsmonsen"
)

var (
	qd   *quoteshuffle.QuoteData
	re   = regexp.MustCompile(pattern)
	_log = log.With().Str("plugin", plugin).Logger()
)

func larsmonsen(command *bot.PassiveCmd) (string, error) {
	if re.MatchString(command.Raw) {
		return qd.Shuffle(), nil
	}
	return "", nil
}

func init() {
	var err error
	qd, err = quoteshuffle.New(util.EnvDefStr("LARSMONSENFACTS_FILE", factsFile))
	if err != nil {
		_log.Error().
			Err(err).
			Msg("Error loading facts file")
		return
	}
	bot.RegisterPassiveCommand(cmdName, larsmonsen)
}
