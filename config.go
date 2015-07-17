package main

import (
	"fmt"
	log "gopkg.in/inconshreveable/log15.v2"
	"time"
)

type Config struct {
	Brokers  []Broker
	LoggerID int64
}

// This handles the timing of re-checking configuration through
// a Go channel that is never closed, hence endlessly repeating
func configTimer(interval *string) chan struct{} {

	go loadBrokers()

	dur, err := time.ParseDuration(*interval)
	if err != nil {
		dur = 1 * time.Minute
	}

	log.Info(fmt.Sprintf("Re-loading configuration every %s", dur))

	ticker := time.NewTicker(dur)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info("Re-checking/loading brokers config")
				loadBrokers()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit

}
