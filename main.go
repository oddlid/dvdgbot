package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-chat-bot/bot/irc"
	_ "github.com/oddlid/dvdgbot/larsmonsen"
	"github.com/oddlid/dvdgbot/leet"
	_ "github.com/oddlid/dvdgbot/morse"
	_ "github.com/oddlid/dvdgbot/timestamp"
	_ "github.com/oddlid/dvdgbot/xkcdbot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	//"github.com/oddlid/dvdgbot/userwatch"
)

const (
	DEF_ADDR string = "irc.oftc.net:6697"
	DEF_USER string = "leetbot"
	DEF_NICK string = "leetbot"
)

var (
	COMMIT_ID  string
	BUILD_DATE string
	VERSION    string
	BIN_NAME   string
)

//func envDefStr(key, fallback string) string {
//	val, found := os.LookupEnv(key)
//	if !found {
//		return fallback
//	}
//	return val // might still be empty, if set, but empty in ENV
//}

func entryPoint(ctx *cli.Context) error {
	c := &irc.Config{
		Channels: ctx.StringSlice("channel"),
		Server:   ctx.String("server"),
		User:     ctx.String("user"),
		Nick:     ctx.String("nick"),
		Password: ctx.String("password"),
		UseTLS:   ctx.Bool("tls"),
		Debug:    ctx.Bool("debug"),
	}

	// If using leet, but not userwatch, do this:
	b, _ := irc.SetUpConn(c) // ic should be second return param here if using userwatch module
	leet.SetParentBot(b)

	// Or, if using both leet and userwatch, do like this instead, and comment the above:
	//b, ic := irc.SetUpConn(c)
	//err := userwatch.InitBot(c, b, ic, envDefStr("USERWATCH_CFGFILE", userwatch.DEF_CFGFILE))
	//if err != nil {
	//	return cli.NewExitError(err.Error(), 1)
	//}
	//leet.SetParentBot(b)

	irc.Run(nil) // pass nil here, as we passed c to SetUpConn, so config is done

	// If not using neither leet nor userwatch, you can comment out both ways to setup above,
	// including the line with "irc.Run(nil)", and replace it with the below:
	//irc.Run(c)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = BIN_NAME
	app.Version = fmt.Sprintf("%s_%s (Compiled: %s)", VERSION, COMMIT_ID, BUILD_DATE)
	app.Compiled, _ = time.Parse(time.RFC3339, BUILD_DATE)
	app.Copyright = fmt.Sprintf("(C) 2018 - %d, Odd Eivind Ebbesen", time.Now().Year())
	app.Authors = []*cli.Author{
		{
			Name:  "Odd E. Ebbesen",
			Email: "oddebb@gmail.com",
		},
	}
	app.Usage = "Run irc bot"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "IRC server `address`",
			Value:   DEF_ADDR,
			EnvVars: []string{"IRC_SERVER"},
		},
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "IRC `username`",
			Value:   DEF_USER,
			EnvVars: []string{"IRC_USER"},
		},
		&cli.StringFlag{
			Name:    "nick",
			Aliases: []string{"n"},
			Usage:   "IRC `nick`",
			Value:   DEF_NICK,
			EnvVars: []string{"IRC_NICK"},
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"p"},
			Usage:   "IRC server `password`",
			EnvVars: []string{"IRC_PASS"},
		},
		&cli.StringSliceFlag{
			Name:    "channel",
			Aliases: []string{"c"},
			Usage:   "Channel to join. May be repeated. Specify \"#chan passwd\" if a channel needs a password.",
		},
		&cli.BoolFlag{
			Name:    "tls",
			Aliases: []string{"t"},
			Usage:   "Use secure TLS connection",
			Value:   true,
			EnvVars: []string{"IRC_TLS"},
		},
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Value:   "info",
			Usage:   "Log `level` (options: debug, info, warn, error, fatal, panic)",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "Run in debug mode",
			EnvVars: []string{"DEBUG"},
		},
	}
	app.Before = func(c *cli.Context) error {
		zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999-07:00"
		if c.Bool("debug") {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			if c.IsSet("log-level") || c.IsSet("l") {
				level, err := zerolog.ParseLevel(c.String("log-level"))
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
		//slog := zerolog.New(os.Stdout).With().Logger()
		//stdlog.SetFlags(0)
		//stdlog.SetOutput(slog)

		return nil
	}

	app.Action = entryPoint
	err := app.Run(os.Args)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
