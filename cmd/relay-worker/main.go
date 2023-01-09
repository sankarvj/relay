package main

import (
	"context"
	"expvar"
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
			APIHost         string        `conf:"default:0.0.0.0:3001,env:WEB_API_HOST""`
			DebugHost       string        `conf:"default:0.0.0.0:4001,env:WEB_DEBUG_HOST"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres,env:DB_USER"`
			Password   string `conf:"default:postgres,noprint,env:DB_PASSWORD"`
			Host       string `conf:"default:0.0.0.0,env:DB_HOST"`
			Name       string `conf:"default:relaydb,env:DB_NAME"`
			DisableTLS bool   `conf:"default:true"`
		}
		SecDB struct {
			User     string `conf:"default:redisgraph,env:SEC_DB_USER"`
			Password string `conf:"default:redis,noprint,env:SEC_DB_PASSWORD"`
			Host     string `conf:"default:127.0.0.1:6379,env:SEC_DB_HOST"`
			Name     string `conf:"default:relaydb,env:SEC_DB_NAME"`
		}
		CacheDB struct {
			User     string `conf:"default:redisgraph,env:CACHE_DB_USER"`
			Password string `conf:"default:redis,noprint,env:CACHE_DB_PASSWORD"`
			Host     string `conf:"default:127.0.0.1:6379,env:CACHE_DB_HOST"`
			Name     string `conf:"default:relaydb,env:CACHE_DB_NAME"`
		}
		Auth struct {
			KeyID              string `conf:"default:1"`
			PrivateKeyFile     string `conf:"default:private.pem,env:AUTH_PRIVATE_KEY_FILE"`
			Algorithm          string `conf:"default:RS256"`
			GoogleKeyFile      string `conf:"default:config/dev/relay-70013-firebase-adminsdk-cfun3-58caec85f0.json,env:AUTH_GOOGLE_KEY_FILE"`
			GoogleClientSecret string `conf:"default:config/dev/google-apps-client-secret.json,env:AUTH_GOOGLE_CLIENT_SECRET"`
		}
		Service struct {
			Region       string `conf:"default:us-east-1,env:AWS_REGION"`
			WorkerSqsURL string `conf:"default:us-east-1,env:AWS_WORKER_SQS_URL"`
		}
		Payment struct {
			StripeLiveKey    string `conf:"default:sk_test_51M0BSXHUBFGeRHv5Qalelfhv8NO1kdnM0FgGd37iG74b2HNQfRLSolOgcvuFjvkfRP4KYTmZwztk5qMCmN245IDW00IUDFBOmp,env:STRIPE_LIVE_KEY"`
			StripePublishKey string `conf:"default:whsec_41d7022cc154e767fe96054ac413c1cde21b2d9c23b4c7743f20315901f247cc,env:STRIPE_PUBLISH_KEY"`
		}
		Build string `conf:"default:dev,env:BUILD"`
	}

	if err := conf.Parse(os.Args[1:], "WORKER", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("WORKER", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "error: parsing config")
	}

	// Store Global Variables
	expvar.NewString("build").Set(cfg.Build)
	expvar.NewString("aws_region").Set(cfg.Service.Region)
	expvar.NewString("aws_worker_sqs_url").Set(cfg.Service.WorkerSqsURL)
	expvar.NewString("stripe_live_key").Set(cfg.Payment.StripeLiveKey)
	expvar.NewString("stripe_publish_key").Set(cfg.Payment.StripePublishKey)

	// =========================================================================
	// App Starting
	log.Printf("main : Started : Application initializing : version %q", cfg.Build)
	defer log.Println("main : Completed")

	// This is used for multiple commands below.
	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	db, err := database.Open(dbConfig)
	if err != nil {
		return errors.Wrap(err, "connecting to primary db")
	}
	defer func() {
		log.Printf("main : Primary Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	path := listeners.Path{
		FirebaseSDKPath: cfg.Auth.GoogleKeyFile,
	}

	secDbConfig := database.SecConfig{
		User:     cfg.SecDB.User,
		Password: cfg.SecDB.Password,
		Host:     cfg.SecDB.Host,
		Name:     cfg.SecDB.Name,
	}

	rp := &redis.Pool{
		MaxIdle:     50,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", secDbConfig.Host, redis.DialPassword(secDbConfig.Password))
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

	// =========================================================================
	// Initialize cache database

	log.Println("main : Started : Initializing cache database support")
	cp := &redis.Pool{
		MaxIdle:     50,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.CacheDB.Host, redis.DialPassword(cfg.CacheDB.Password))
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
		log.Printf("main : Redis Cache Database Stopping : %s", cfg.CacheDB.Host)
		rp.Close()
	}()

	sdb := database.Init(rp, cp, rp)

	//this should be started as the separate service.
	go func() {
		log.Printf("main : Debug Running Job Listener")
		l := job.Listener{}
		l.RunReminderListener(db, sdb, path.FirebaseSDKPath)
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

	handler := c.Handler(listeners.API(shutdown, log, db, sdb, path))

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
