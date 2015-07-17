package main

import (
	"fmt"
	log "gopkg.in/inconshreveable/log15.v2"
)

// This function logs a fatal error and kills the process.
func fatalErr(err error, prefix string) {
	if err != nil {
		defer db.Close()
		msg := fmt.Sprintf("%s: %s", prefix, err.Error())
		log.Error(msg)
		panic(msg)
	}
}

// Logs a non-fatal error.
func checkErr(err error, prefix string) {
	if err != nil {
		msg := fmt.Sprintf("%s: %s", prefix, err.Error())
		log.Error(msg)
	}
}
