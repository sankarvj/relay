package main

import (
	"log"
	"os"

	"gitlab.com/vjsideprojects/relay/internal/job"
)

func main() {
	if err := run(); err != nil {
		log.Printf("main : error: %s", err)
		os.Exit(1)
	}
}

func run() error {
	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "RELAY WORKER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration

	//this should be started as the separate service.
	go func() {
		log.Printf("main : Debug Running Worker")
		l := job.Listener{}
		l.RunReminderListener(redisPool)
	}()

	return nil
}
