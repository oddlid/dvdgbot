package xkcdbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oddlid/bot"
	// Do NOT run goimports on this file, as it will remove imports where the last part of path does not match pkg name
	// Use gofmt instead!
	"github.com/nishanths/go-xkcd"
)

var (
	xc *xkcd.Client
)

func xkcdbot(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return "Too few params. Usage: !xkcd get <ID>|random|latest", nil
	}

	switch strings.ToUpper(cmd.Args[0]) {
	case "GET":
		if len(cmd.Args) < 2 {
			return "Parameter GET needs an ID", nil
		}
		id, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return "ID for GET must be a number", nil
		}
		comic, err := xc.Get(id)
		if err != nil {
			return fmt.Sprintf("Error fetching ID #%d", id), nil
		}
		return comic.ImageURL, nil
	case "RANDOM":
		comic, err := xc.Random()
		if err != nil {
			return "Error fetching random comic", nil
		}
		return comic.ImageURL, nil
	case "LATEST":
		comic, err := xc.Latest()
		if err != nil {
			return "Error fetching latest comic", nil
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
		xkcdbot)
}
