package main

import (
	"flag"
	"fmt"
	"os"
)

// Config is the configuration of the CLI
type Config struct {
	Debug   bool
	GetSize bool
	ID      string
}

// Initial value of the config
var cfg = &Config{
	false,
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
	fmt.Println(fmt.Sprintf("Guide: %s [--debug --id --getsize]", os.Args[0]))
	fmt.Println(fmt.Sprintf("Example: %s --id 3135556", os.Args[0]))
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&cfg.Debug, "debug", false, "Turn on debuging mode.")
	flag.BoolVar(&cfg.GetSize, "getsize", false, "Only Get the Size of the 320kpbs audio")
	flag.StringVar(&cfg.ID, "id", "", "Deezer Track ID")

	flag.Parse()

	// fmt.Println("Make Sure You Register your Deezer")
	debug("Configuration:")

	debug("\tDebug: %t", cfg.Debug)
	debug("\tID: %s", cfg.ID)

	if cfg.ID == "" {
		fmt.Println("Error: Must have Deezer Track(Song) ID")
		ErrorUsage()
	}
}
