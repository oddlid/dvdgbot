package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oddlid/bot"
	"github.com/oddlid/bot/irc"
	_ "github.com/oddlid/dvdgbot/larsmonsen"
	_ "github.com/oddlid/dvdgbot/leet"
	"github.com/oddlid/dvdgbot/userwatch"
	_ "github.com/oddlid/dvdgbot/xkcdbot"
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

func entryPoint(ctx *cli.Context) error {
	tls := ctx.Bool("tls")
	dbg := ctx.Bool("debug")
	serv := ctx.String("server")
	nick := ctx.String("nick")
	user := ctx.String("user")
	pass := ctx.String("password")
	chans := ctx.StringSlice("channel")

	log.WithFields(log.Fields{
		"server":   serv,
		"password": pass,
		"user":     user,
		"nick":     nick,
		"tls":      tls,
		"channels": chans,
	}).Debug("Received config:")
	//	if dbg {
	//		return nil
	//	}

	c := &irc.Config{
		Server:   serv,
		Channels: chans,
		User:     user,
		Nick:     nick,
		Password: pass,
		UseTLS:   tls,
		Debug:    dbg,
	}

	bot.UseUnidecode = false // my own hack so we can have nice scandinavian letters
	b, ic := irc.SetUpConn(c)
	err := userwatch.InitBot(c, b, ic, userwatch.DEF_CFGFILE)
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
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatal(err.Error())
		}
		log.SetLevel(level)
		if !c.IsSet("log-level") && !c.IsSet("l") && c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: false,
			FullTimestamp:    true,
		})
		return nil
	}

	app.Action = entryPoint
	app.Run(os.Args)
}
