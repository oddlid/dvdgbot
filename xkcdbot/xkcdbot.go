package xkcdbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-chat-bot/bot"
	// Do NOT run goimports on this file, as it will remove imports where the last part of path does not match pkg name
	// Use gofmt instead!
	"github.com/nishanths/go-xkcd/v2"
)

var xc *xkcd.Client

func xkcdbot(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return "Too few params. Usage: !xkcd get <ID>|latest", nil
	}

	switch strings.ToUpper(cmd.Args[0]) {
	case "GET":
		if len(cmd.Args) < 2 {
			return "Parameter GET needs an ID", nil
		}
		id, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return "ID for GET must be a number", err
		}
		comic, err := xc.Get(context.Background(), id)
		if err != nil {
			return fmt.Sprintf("Error fetching ID #%d", id), err
		}
		return comic.ImageURL, nil
	// the Random() method seems to have disappeared in v2
	// case "RANDOM":
	//	comic, err := xc.Random()
	//	if err != nil {
	//		return "Error fetching random comic", nil
	//	}
	//	return comic.ImageURL, nil
	case "LATEST":
		comic, err := xc.Latest(context.Background())
		if err != nil {
			return "Error fetching latest comic", err
		}
		return comic.ImageURL, nil
	}

	return "", nil
}

func init() {
	xc = xkcd.NewClient()
	bot.RegisterCommand(
		"xkcd",
		"Fetch an XKCD comic image",
		"get <ID>|random|latest",
		xkcdbot,
	)
}
