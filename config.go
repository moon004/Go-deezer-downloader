package main

import (
	"flag"
	"fmt"
	"os"
)

// Config is the configuration of the CLI
type Config struct {
	Debug     bool
	File      bool
	ID        string
	UserToken string
}

// Initial value of the config
var cfg = &Config{
	false,
	false,
	"",
	"",
}

func debug(msg string, params ...interface{}) {
	if cfg.Debug {
		fmt.Printf("\n"+msg+"\n\n", params...)
	}
}

// ErrorUsage lets the user knows the error
func ErrorUsage() {
	fmt.Println(`Guide: go-decrypt-deezer [--debug --id --usertoken`)
	fmt.Println(`How Do I Get My UserToken?: https://notabug.org/RemixDevs/DeezloaderRemix/wiki/Login+via+userToken`)
	fmt.Println(`Example: go-decrypt-deezer --id 3135556 --usertoken UserToken_here`)
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&cfg.Debug, "debug", false, "Turn on debuging mode.")
	flag.BoolVar(&cfg.File, "f", false, "output to file")
	flag.StringVar(&cfg.UserToken, "usertoken", "", "Your Unique User Token")
	flag.StringVar(&cfg.ID, "id", "", "Deezer Track ID")

	flag.Parse()

	// fmt.Println("Make Sure You Register your Deezer")
	debug("Configuration:")

	debug("\tDebug: %t", cfg.Debug)
	debug("\tUserToken: %s", cfg.UserToken)
	debug("\tID: %s", cfg.ID)

	if cfg.ID == "" {
		fmt.Println("Error: Must have Deezer Track(Song) ID")
		ErrorUsage()
	}

}
