package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chat-bot/bot/irc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	_ "github.com/oddlid/dvdgbot/larsmonsen"
	"github.com/oddlid/dvdgbot/leet"
	_ "github.com/oddlid/dvdgbot/morse"
	_ "github.com/oddlid/dvdgbot/timestamp"
	_ "github.com/oddlid/dvdgbot/xkcdbot"
	// "github.com/oddlid/dvdgbot/userwatch"
)

const (
	defaultAddress string = "irc.oftc.net:6697"
	defaultUser    string = "leetbot"
	defaultNick    string = "leetbot"
)

const (
	optServer   = `server`
	optChannel  = `channel`
	optUser     = `user`
	optNick     = `nick`
	optPass     = `password`
	optTLS      = `tls`
	optDebug    = `debug`
	optLogLevel = `log-level`
)

var (
	CommitID   string
	BuildDate  string
	Version    string
	BinaryName string
)

func entryPoint(ctx *cli.Context) error {
	c := &irc.Config{
		Channels: ctx.StringSlice(optChannel),
		Server:   ctx.String(optServer),
		User:     ctx.String(optUser),
		Nick:     ctx.String(optNick),
		Password: ctx.String(optPass),
		UseTLS:   ctx.Bool(optTLS),
		Debug:    ctx.Bool(optDebug),
	}

	// If using leet, but not userwatch, do this:
	b, _ := irc.SetUpConn(c) // ic should be second return param here if using userwatch module
	leet.SetParentBot(b)

	// Or, if using both leet and userwatch, do like this instead, and comment the above:
	// b, ic := irc.SetUpConn(c)
	// err := userwatch.InitBot(c, b, ic, envDefStr("USERWATCH_CFGFILE", userwatch.DEF_CFGFILE))
	// if err != nil {
	// 	return cli.NewExitError(err.Error(), 1)
	// }
	// leet.SetParentBot(b)

	irc.Run(nil) // pass nil here, as we passed c to SetUpConn, so config is done

	// If not using neither leet nor userwatch, you can comment out both ways to setup above,
	// including the line with "irc.Run(nil)", and replace it with the below:
	// irc.Run(c)

	return nil
}

func parseTime(in string) time.Time {
	if t, err := time.Parse(time.RFC3339, in); err == nil {
		return t
	}
	return time.Time{}
}

func newApp() *cli.App {
	return &cli.App{
		Name:      BinaryName,
		Version:   fmt.Sprintf("%s_%s (Compiled: %s)", Version, CommitID, BuildDate),
		Compiled:  parseTime(BuildDate),
		Copyright: fmt.Sprintf("(C) 2018 - %d, Odd Eivind Ebbesen", time.Now().Year()),
		Usage:     "Run irc bot",
		Authors: []*cli.Author{
			{
				Name:  "Odd E. Ebbesen",
				Email: "oddebb@gmail.com",
			},
		},
		Before: func(c *cli.Context) error {
			zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999-07:00"
			if c.Bool(optDebug) {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				if c.IsSet(optLogLevel) || c.IsSet("l") {
					level, err := zerolog.ParseLevel(c.String(optLogLevel))
					if err != nil {
						log.Error().Err(err).Send()
					} else {
						zerolog.SetGlobalLevel(level)
					}
				} else {
					zerolog.SetGlobalLevel(zerolog.InfoLevel)
				}
			}
			// I had hoped this could wrap the logging from the underlying libs, but it seems they log
			// with just fmt.Println() or something, so this does nothing
			// slog := zerolog.New(os.Stdout).With().Logger()
			// stdlog.SetFlags(0)
			// stdlog.SetOutput(slog)

			return nil
		},
		Action: entryPoint,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    optLogLevel,
				Aliases: []string{"s"},
				Usage:   "IRC server `address`",
				Value:   defaultAddress,
				EnvVars: []string{"IRC_SERVER"},
			},
			&cli.StringFlag{
				Name:    optUser,
				Aliases: []string{"u"},
				Usage:   "IRC `username`",
				Value:   defaultUser,
				EnvVars: []string{"IRC_USER"},
			},
			&cli.StringFlag{
				Name:    optNick,
				Aliases: []string{"n"},
				Usage:   "IRC `nick`",
				Value:   defaultNick,
				EnvVars: []string{"IRC_NICK"},
			},
			&cli.StringFlag{
				Name:    optPass,
				Aliases: []string{"p"},
				Usage:   "IRC server `password`",
				EnvVars: []string{"IRC_PASS"},
			},
			&cli.StringSliceFlag{
				Name:    optChannel,
				Aliases: []string{"c"},
				Usage:   "Channel to join. May be repeated. Specify \"#chan passwd\" if a channel needs a password.",
			},
			&cli.BoolFlag{
				Name:    optTLS,
				Aliases: []string{"t"},
				Usage:   "Use secure TLS connection",
				Value:   true,
				EnvVars: []string{"IRC_TLS"},
			},
			&cli.StringFlag{
				Name:    optLogLevel,
				Aliases: []string{"l"},
				Value:   "info",
				Usage:   "Log `level` (options: debug, info, warn, error, fatal, panic)",
			},
			&cli.BoolFlag{
				Name:    optDebug,
				Aliases: []string{"d"},
				Usage:   "Run in debug mode",
				EnvVars: []string{"DEBUG"},
			},
		},
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	if err := newApp().RunContext(ctx, os.Args); err != nil {
		log.Error().Err(err).Send()
	}
}
