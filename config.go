package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Debug    bool
	Username string
	Password string
	ID       string
}

var cfg *Config = &Config{ //default values
	false,
	"",
	"",
	"",
}

func debug(msg string, params ...interface{}) {
	if cfg.Debug {
		fmt.Printf("\n"+msg+"\n", params...)
	}
}

func error_usage() {
	fmt.Println(`Guide: go-decrypt-deezer [--debug --id --username --password]`)
	fmt.Println(`Example: go-decrypt-deezer --id 3135556 --username username_here --password password_here`)
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&cfg.Debug, "debug", false, "Turn on debuging mode.")
	flag.StringVar(&cfg.Username, "username", "", "Your Deezer Username")
	flag.StringVar(&cfg.Password, "password", "", "Your Deezer Password")
	flag.StringVar(&cfg.ID, "id", "", "Deezer Track ID")

	flag.Parse()

	// fmt.Println("Make Sure You Register your Deezer")
	debug("Configuration:")

	debug("\tDebug: %t", cfg.Debug)
	debug("\tUsername: %s", cfg.Username)
	debug("\tPassword: %s", cfg.Password)
	debug("\tID: %s", cfg.ID)

	if cfg.ID == "" {
		fmt.Println("Error: Must have Deezer Track ID")
		error_usage()
	}
	if cfg.Username == "" {
		fmt.Println("Error: Must have Username (Mail)")
		error_usage()
	}
	if cfg.Password == "" {
		fmt.Println("Error: Must have Password")
		error_usage()
	}

}
