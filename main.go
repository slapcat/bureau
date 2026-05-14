package main

import (
	"os"
	"time"
)

var C Config

func main() {

	// load bureau config
	C, err := ConfigInit()
	Logger(err, "Error reading config file", "FATAL")
	
	for {

		// run daemon once
		Summon()

		// Cleanup
		Tpl = nil

		// loop if in daemon mode
		if C.Daemon {
			time.Sleep(time.Duration(C.Update) * time.Second)
		} else {
			os.Exit(0)
		}

	}
}
