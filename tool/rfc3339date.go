//go:build ignore

/*
This file exists only because ´date´ in OSX can't print
a proper RFC3339 date, so I made this hack.
*/

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now().Format(time.RFC3339))
}
