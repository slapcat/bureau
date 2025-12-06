package main

import (
	"os"
	"time"
)

var c Config

func main() {

	// load bureau config
	c, err := ConfigInit()
	Logger(err, "Error reading config file", "FATAL")

	for {

		// run daemon once
		Summon()

		// Cleanup
		Tpl = nil

		// loop if in daemon mode
		if c.Daemon {
			time.Sleep(time.Duration(c.Update) * time.Second)
		} else {
			os.Exit(0)
		}

	}
}
