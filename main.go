package main

import (
	"log"
	"os"
	"time"
)

var c Config
var f File
var k Kalived

func main() {

	// load bureau config
	c, err := ConfigInit()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	for {

		// run daemon once
		Summon()

		// loop if in daemon mode
		if c.Daemon {
			time.Sleep(time.Duration(c.Update) * time.Second)
		} else {
			os.Exit(0)
		}

	}
}
