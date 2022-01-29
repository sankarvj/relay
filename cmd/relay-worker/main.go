package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"gitlab.com/vjsideprojects/relay/cmd/relay-worker/internal/listeners"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
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

	var cfg struct {
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3001"`
			DebugHost       string        `conf:"default:0.0.0.0:4001"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:0.0.0.0"`
			Name       string `conf:"default:relaydb"`
			DisableTLS bool   `conf:"default:true"`
		}
		SecDB struct {
			User     string `conf:"default:redisgraph"`
			Password string `conf:"default:redis,noprint"`
			Host     string `conf:"default:127.0.0.1:6379"`
			Name     string `conf:"default:relaydb"`
		}
		Args conf.Args
	}

	if err := conf.Parse(os.Args[1:], "RELAY WORKER", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("RELAY WORKER", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "error: parsing config")
	}

	// This is used for multiple commands below.
	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	secDbConfig := database.SecConfig{
		User:     cfg.SecDB.User,
		Password: cfg.SecDB.Password,
		Host:     cfg.SecDB.Host,
		Name:     cfg.SecDB.Name,
	}

	db, err := database.Open(dbConfig)
	if err != nil {
		return errors.Wrap(err, "connecting to primary db")
	}
	defer func() {
		log.Printf("main : Primary Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	rp := &redis.Pool{
		MaxIdle:     50,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", secDbConfig.Host)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	defer func() {
		log.Printf("main : Redis Database Stopping : %s", cfg.SecDB.Host)
		rp.Close()
	}()

	//this should be started as the separate service.
	go func() {
		log.Printf("main : Debug Running Job Listener")
		l := job.Listener{}
		l.RunReminderListener(db, rp)
	}()

	// =========================================================================
	// Start WebServer

	go func() {
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "GET", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "X-Requested-With", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(listeners.API(shutdown, log, db, rp))

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handler,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}