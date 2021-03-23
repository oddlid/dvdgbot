# leet / 1337

This is a game where the point is to post `!1337` as close to 13:37:00:000000000 as possible, on the positive side.

## Installation

Import `github.com/oddlid/dvdgbot/leet` in `main.go`. Then, in `entryPoint()` in `main.go`, take the first return value from `irc.SetUpConn(config)` (the bot instance), and pass on to `leet.SetParentBot(bot)`.

## Config

The config for the game is done via both environment variables and in the JSON config files it needs.

### Enviroment variables:

* `LEETBOT_HOUR` - Defaults to 13. Useful to change to current hour during tests.
* `LEETBOT_MINUTE` - Defaults to 37. Useful to change to current minute during tests.
* `LEETBOT_SCOREFILE` - Defaults to `/tmp/leetbot_scores.json`. This is where The config for channels and their respective settings goes, and where the bot saves scores, times, tax and bonuses for each user.
* `LEETBOT_BONUSCONFIGFILE` - Defaults to `/tmp/leetbot_bonusconfigs.json`. This is where you configure the bonus system. The bonus system is based on substring matching in the second and nanosecond fields of the timestamp when a user's post is registered.

### JSON files:

* `$LEETBOT_SCOREFILE ( /tmp/leetbot_scores.json )` - Main configuration. Format:
```
{
  "botstart": "timestamp automatically set by bot",
  "channels": {
    "#channelname": {
      "channel_name": "#channelname",
      "users": {
        "Nick1": {
          "nick": "Nick1",
          "score": 0,
          "last_entry": "timestamp automatically set by bot",
          "best_entry": "timestamp automatically set by bot",
          "locked": false,
          "bonuses": {
            "times": 0,
            "total": 0
          },
          "taxes": {
            "times": 0,
            "total": 0
          }
        },
        "Nick2": { ... }
      },
      "inspect_always": false,
      "tax_loners": false,
      "post_tax_fail": false,
      "inspection_tax": 0,
      "overshoot_tax": 0
    }
  }
}
```

Most of the data in this file is generated automatically by the bot itself, used for keeping track of stats between restarts. The only values you should touch, are the ones at the bottom:

* `inspect_always`: true/false
  - If set to true, Tax Inspection will be run after every round.
  - If set to false, Tax Inspection will only be run if a random int between 0 and 6 matches the current weekday.
* `tax_loners`: true/false
  - If there is only one contestant on time in a given round, Tax Inspection will not run if this is set to `false`. `inspect_always` will override this, if set to `true`, though.
* `post_tax_fail`: true/false
  - If set to `true`, the bot will post why taxation was not done to the channel. Usually, it's because the weekday doesn't match the random int.
* `inspection_tax`: int
  - This is a percentage of the lowest score between the contestants that scored on time in a given round. So if the contestant with the higest total points have 1000 points, and the one with the lowest has 200, and `inspection_tax` is set to 10, then 20 points is the maximum tax for that round.
* `overshoot_tax`: int
  - This is how many points will be the step value to deduct in a loop until a user is below the target score, if scoring past that value. The target score will be the concatenation of `LEETBOT_HOUR` and `LEETBOT_MINUTE`, so if set to defaults, the target score will be 1337 points. So if the overshoot tax is 10, a user has 1336 points, and gets 2 points in a round, it will be deducted 10 points and have 1328 points after the round. If the user overshoots with more points than the value of this tax, it will be decuted in a loop until the value is below.
