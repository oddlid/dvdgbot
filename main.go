package main

import (
	"fmt"
	stdLog "log"
	"os"
	"time"

	"github.com/go-chat-bot/bot/irc"
	_ "github.com/oddlid/dvdgbot/larsmonsen"
	"github.com/oddlid/dvdgbot/leet"
	_ "github.com/oddlid/dvdgbot/timestamp"
	"github.com/oddlid/dvdgbot/userwatch"
	_ "github.com/oddlid/dvdgbot/xkcdbot"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	//_ "github.com/go-chat-bot/plugins/chucknorris"
	//_ "github.com/oddlid/dvdgbot/goodmorning"
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
)

func envDefStr(key, fallback string) string {
	val, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return val // might still be empty, if set, but empty in ENV
}

func entryPoint(ctx *cli.Context) error {
	c := &irc.Config{
		Server:   ctx.String("server"),
		Channels: ctx.StringSlice("channel"),
		User:     ctx.String("user"),
		Nick:     ctx.String("nick"),
		Password: ctx.String("password"),
		UseTLS:   ctx.Bool("tls"),
		Debug:    ctx.Bool("debug"),
	}

	b, ic := irc.SetUpConn(c)
	leet.SetParentBot(b)
	err := userwatch.InitBot(c, b, ic, envDefStr("USERWATCH_CFGFILE", userwatch.DEF_CFGFILE))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	irc.Run(nil) // pass nil here, as we passed c to SetUpConn

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "bajsbot"
	app.Version = fmt.Sprintf("%s_%s (Compiled: %s)", VERSION, COMMIT_ID, BUILD_DATE)
	app.Compiled, _ = time.Parse(time.RFC3339, BUILD_DATE)
	app.Copyright = "(c) 2018 Odd Eivind Ebbesen"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Odd E. Ebbesen",
			Email: "oddebb@gmail.com",
		},
	}
	app.Usage = "Run irc bot"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "server, s",
			Usage:  "IRC server `address`",
			Value:  DEF_ADDR,
			EnvVar: "IRC_SERVER",
		},
		cli.StringFlag{
			Name:   "user, u",
			Usage:  "IRC `username`",
			Value:  DEF_USER,
			EnvVar: "IRC_USER",
		},
		cli.StringFlag{
			Name:   "nick, n",
			Usage:  "IRC `nick`",
			Value:  DEF_NICK,
			EnvVar: "IRC_NICK",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "IRC server `password`",
			EnvVar: "IRC_PASS",
		},
		cli.StringSliceFlag{
			Name:  "channel, c",
			Usage: "Channel to join. May be repeated. Specify \"#chan passwd\" if a channel needs a password.",
		},
		cli.BoolTFlag{
			Name:   "tls, t",
			Usage:  "Use secure TLS connection",
			EnvVar: "IRC_TLS",
		},
		cli.StringFlag{
			Name:  "log-level, l",
			Value: "info",
			Usage: "Log `level` (options: debug, info, warn, error, fatal, panic)",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Run in debug mode",
			EnvVar: "DEBUG",
		},
	}
	app.Before = func(c *cli.Context) error {
		//log.SetOutput(os.Stderr) // this is the default anyways, from Logrus package
		if !c.IsSet("log-level") && !c.IsSet("l") && c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		} else {
			level, err := log.ParseLevel(c.String("log-level"))
			if err != nil {
				log.Fatal(err.Error())
			}
			log.SetLevel(level)
		}
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: false,
			FullTimestamp:    true,
		})

		// Overwrite STD logger used in foreign packages
		stdLog.SetOutput(log.StandardLogger().WriterLevel(log.GetLevel()))
		// Or:
		//stdLog.SetOutput(log.StandardLogger().Writer())

		return nil
	}

	app.Action = entryPoint
	app.Run(os.Args)
}
