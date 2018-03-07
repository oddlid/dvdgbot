package xkcdbot

import (
	"github.com/go-chat-bot/bot"
	xkcd "github.com/nishanths/go-xkcd"
)

var (
	xc xkcd.Client
)

func xkcd(cmd *bot.Cmd) (string, error) {
	comic, err := xc.Random()
	return comic.ImageURL, nil
}

func init() {
	xc = xkcd.NewClient()
	bot.RegisterCommand(
		"xkcd",
		"Fetch an XKCD comic image",
		"get <ID>|random|latest",
		xkcd)
}
