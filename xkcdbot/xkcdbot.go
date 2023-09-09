package xkcdbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-chat-bot/bot"
	"github.com/nishanths/go-xkcd/v2"
)

const (
	DefaultCommandName = `xkcd`
	Description        = `Fetch an XKCD comic image`
	Params             = `get <ID>|random|latest`
)

type ContextFunc func() context.Context

type Bot struct {
	client  *xkcd.Client
	ctxFunc ContextFunc
	timeout time.Duration
}

func New(timeout time.Duration, ctxFunc ContextFunc) *Bot {
	return &Bot{
		client:  xkcd.NewClient(),
		ctxFunc: ctxFunc,
		timeout: timeout,
	}
}

func (b *Bot) Fetch(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return "Too few params. Usage: !xkcd get <ID>|latest", nil
	}

	ctx, cancel := context.WithTimeout(b.ctxFunc(), b.timeout)
	defer cancel()

	switch strings.ToUpper(cmd.Args[0]) {
	case "GET":
		if len(cmd.Args) < 2 {
			return "Parameter GET needs an ID", nil
		}
		id, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return "ID for GET must be a number", err
		}
		comic, err := b.client.Get(ctx, id)
		if err != nil {
			return fmt.Sprintf("Error fetching ID #%d: %s", id, err.Error()), err
		}
		return comic.ImageURL, nil
	case "LATEST":
		comic, err := b.client.Latest(ctx)
		if err != nil {
			return fmt.Sprintf("Error fetching latest comic: %s", err.Error()), err
		}
		return comic.ImageURL, nil
	default:
		return "", nil
	}
}
