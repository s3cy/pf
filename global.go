package main

import (
	"log"
	"os"
	"os/user"
)

var (
	UserHomeDir string
	UserName    string
	HostName    string
	StyleM      StyleMap
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	UserHomeDir = home

	user, err := user.Current()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	UserName = user.Name

	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	HostName = host

	StyleM = ParseStyles()
}
