package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func main() {
	if err := run(); err != nil {
		log.Printf("main : error: %s", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Configuration

	var cfg struct {
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

	if err := conf.Parse(os.Args[1:], "RELAY ADMIN", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("RELAY ADMIN", &cfg)
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

	switch cfg.Args.Num(0) {
	case "migrate":
		err = migrate(dbConfig)
	case "seed":
		err = seed(db, rp)
	case "crmadd":
		err = bootstrap.BootCRM(schema.SeedAccountID, db, rp)
	case "csmadd":
		err = bootstrap.BootCSM(schema.SeedAccountID, db, rp)
	case "ctmadd":
		err = bootstrap.BootCSM(schema.SeedAccountID, db, rp)
	case "useradd":
		err = useradd(db, cfg.Args.Num(1), cfg.Args.Num(2))
	case "keygen":
		err = keygen(cfg.Args.Num(1))
	default:
		err = errors.New("Must specify a command")
	}

	if err != nil {
		return err
	}

	return nil
}

func migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return err
	}

	fmt.Println("Migrations complete")
	return nil
}

func seed(db *sqlx.DB, rp *redis.Pool) error {

	if err := schema.SeedUsers(db); err != nil {
		return err
	}

	ctx := context.Background()
	cuser, err := user.RetrieveUser(ctx, db, schema.SeedUserID1)
	if err != nil {
		return err
	}
	nc := account.NewAccount{
		ID:     schema.SeedAccountID,
		Name:   "Wayplot",
		Domain: "wayplot.com"}
	err = account.Bootstrap(ctx, db, rp, cuser, nc, time.Now())
	if err != nil {
		log.Println("main: !!!! TODO: Should Implement Roll Back Option Here.")
		return err
	}

	fmt.Println("main: sample data seeded successfully!!!")
	return nil
}

func useradd(db *sqlx.DB, email, password string) error {
	if email == "" || password == "" {
		return errors.New("useradd command must be called with two additional arguments for email and password")
	}

	fmt.Printf("Admin user will be created with email %q and password %q\n", email, password)
	fmt.Print("Continue? (1/0) ")

	var confirm bool
	if _, err := fmt.Scanf("%t\n", &confirm); err != nil {
		return errors.Wrap(err, "processing response")
	}

	if !confirm {
		fmt.Println("Canceling")
		return nil
	}

	ctx := context.Background()

	nu := user.NewUser{
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	u, err := user.Create(ctx, db, nu, time.Now())
	if err != nil {
		return err
	}

	fmt.Println("User created with id:", u.ID)
	return nil
}

// keygen creates an x509 private key for signing auth tokens.
func keygen(path string) error {
	if path == "" {
		return errors.New("keygen missing argument for key path")
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "generating keys")
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating private file")
	}
	defer file.Close()

	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(file, &block); err != nil {
		return errors.Wrap(err, "encoding to private file")
	}

	if err := file.Close(); err != nil {
		return errors.Wrap(err, "closing private file")
	}

	return nil
}
