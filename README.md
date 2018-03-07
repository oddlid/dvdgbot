bajsbot - Just making a stupid irc bot for fun

Currently it has a module for Lars Monsen quotes, based on the Chuck Norris module,
and a "leet" module I wrote from scratch.

The leet module will react to the commands "!1337" and "!1337 stats".

"!1337" will give the nick that enters it 1 point if the time is within 13:37 (you can change the hour and minute with env vars $LEETBOT_HOUR and $LEETBOT_MINUTE). 
If the nick is the first in the channel that day to enter !1337, the nick is awarded 2 points.
If the nick enters !1337 within one minute too early or too late, the nick gets -1 points.
Entering !1337 outside the specified time +- 1 minute, will not do anything.

"!1337 stats" will print the scoreboard since the bot was started. Stats are only kept in memory, and not saved between runs.
"!1337 stats" will work at any given time of day.

See "./bajsbot.bin -h" for options you can give at startup, such as specifying server, channels, user, nick, etc.

