package main

import (
	"flag"
	"fmt"
	"os"
)

// Config is the configuration of the CLI
type Config struct {
	Debug     bool
	ID        string
}

// Initial value of the config
var cfg = &Config{
	false,
	"",
}

func debug(msg string, params ...interface{}) {
	if cfg.Debug {
		fmt.Printf("\n"+msg+"\n\n", params...)
	}
}

// ErrorUsage lets the user knows the error
func ErrorUsage() {
	fmt.Println(`Guide: go-decrypt-deezer [--debug --id]`)
	fmt.Println(`Example: go-decrypt-deezer --id 3135556`)
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&cfg.Debug, "debug", false, "Turn on debuging mode.")
	flag.StringVar(&cfg.ID, "id", "", "Deezer Track ID")

	flag.Parse()
	debug("Configuration:")
	debug("\tDebug: %t", cfg.Debug)
	debug("\tID: %s", cfg.ID)

	if cfg.ID == "" {
		fmt.Println("Error: Must have Deezer Track(Song) ID")
		ErrorUsage()
	}
}
