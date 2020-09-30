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

	"gitlab.com/vjsideprojects/relay/internal/rule/flow"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/config"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s", err)
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

	var err error
	switch cfg.Args.Num(0) {
	case "migrate":
		err = migrate(dbConfig)
	case "seed":
		err = seed(dbConfig)
	case "crmadd":
		err = crmadd(dbConfig)
	case "useradd":
		err = useradd(dbConfig, cfg.Args.Num(1), cfg.Args.Num(2))
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

func seed(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Seed(db); err != nil {
		return err
	}

	// if err := schema.SeedEntity(db); err != nil {
	//    return err
	// }

	fmt.Println("Seed data complete")
	return nil
}

func crmadd(cfg database.Config) error {
	//add entity - status
	se, err := config.EntityAdd(cfg, "Status", entity.CategoryData, config.StatusFields())
	if err != nil {
		return err
	}
	//add entity - contacts
	ce, err := config.EntityAdd(cfg, "Contacts", entity.CategoryData, config.ContactFields(se.ID))
	if err != nil {
		return err
	}
	//add entity - task
	te, err := config.EntityAdd(cfg, "Tasks", entity.CategoryData, config.TaskFields(ce.ID))
	if err != nil {
		return err
	}
	//add entity - deal
	de, err := config.EntityAdd(cfg, "Deals", entity.CategoryData, config.DealFields(ce.ID))
	if err != nil {
		return err
	}
	//add entity - email
	me, err := config.EntityAdd(cfg, "MailGun Intg", entity.CategoryEmail, config.EmailFields())
	if err != nil {
		return err
	}
	//add entity - api-hook
	we, err := config.EntityAdd(cfg, "WebHook", entity.CategoryAPI, config.APIFields())
	if err != nil {
		return err
	}
	//add entity - delay
	dele, err := config.EntityAdd(cfg, "Delay Timer", entity.CategoryDelay, config.DelayFields())
	if err != nil {
		return err
	}

	// add status item - open
	st1, err := config.ItemAdd(cfg, se.ID, config.StatusVals("Open", "#fb667e"))
	if err != nil {
		return err
	}
	// add status item - closed
	st2, err := config.ItemAdd(cfg, se.ID, config.StatusVals("Closed", "#66fb99"))
	if err != nil {
		return err
	}
	// add contact item - vijay (straight)
	con1, err := config.ItemAdd(cfg, ce.ID, config.ContactVals("Vijay", "vijayasankarj@gmail.com", st1.ID))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	con2, err := config.ItemAdd(cfg, ce.ID, config.ContactVals("Senthil", "senthil@gmail.com", st2.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = config.ItemAdd(cfg, te.ID, config.TaskVals("add deal price", con1.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = config.ItemAdd(cfg, te.ID, config.TaskVals("make call", con1.ID))
	if err != nil {
		return err
	}
	// add deal item with contacts - vijay & senthil (reverse)
	_, err = config.ItemAdd(cfg, de.ID, config.DealVals("Big Deal", 1000, con1.ID, con2.ID))
	if err != nil {
		return err
	}
	// add email item
	emg, err := config.ItemAdd(cfg, me.ID, config.EmailVals(ce.ID))
	if err != nil {
		return err
	}
	// add delay item
	delayi, err := config.ItemAdd(cfg, me.ID, config.DelayVals())
	if err != nil {
		return err
	}

	//add workflows
	f, err := config.FlowAdd(cfg, ce.ID, "The Workflow", flow.FlowTypeSegment, flow.FlowConditionEntry)
	if err != nil {
		return err
	}

	//test node push test case - TestCreateItemRuleRunner
	no1, err := config.NodeAdd(cfg, f.ID, te.ID, "", node.Push, "", map[string]string{})
	if err != nil {
		return err
	}

	no2, err := config.NodeAdd(cfg, f.ID, "00000000-0000-0000-0000-000000000000", no1.ID, node.Decision, "{Vijay} eq {Vijay}", map[string]string{})
	if err != nil {
		return err
	}

	no3, err := config.NodeAdd(cfg, f.ID, me.ID, no2.ID, node.Email, "{{xyz.result}} eq {true}", map[string]string{me.ID: emg.ID})
	if err != nil {
		return err
	}

	_, err = config.NodeAdd(cfg, f.ID, we.ID, no2.ID, node.Hook, "{{xyz.result}} eq {false}", map[string]string{})
	if err != nil {
		return err
	}

	_, err = config.NodeAdd(cfg, f.ID, dele.ID, no3.ID, node.Delay, "", map[string]string{dele.ID: delayi.ID})
	if err != nil {
		return err
	}

	p, err := config.FlowAdd(cfg, ce.ID, "The Pipeline", flow.FlowTypePipeline, flow.FlowConditionEntry)
	if err != nil {
		return err
	}

	//test node push test case - TestCreateItemRuleRunner
	pno1, err := config.NodeAdd(cfg, p.ID, "00000000-0000-0000-0000-000000000000", "", node.Stage, "", map[string]string{})
	if err != nil {
		return err
	}

	pno2, err := config.NodeAdd(cfg, p.ID, "00000000-0000-0000-0000-000000000000", pno1.ID, node.Stage, "{Vijay} eq {Vijay}", map[string]string{})
	if err != nil {
		return err
	}

	_, err = config.NodeAdd(cfg, p.ID, me.ID, pno1.ID, node.Email, "", map[string]string{me.ID: emg.ID})
	if err != nil {
		return err
	}

	_, err = config.NodeAdd(cfg, p.ID, we.ID, pno1.ID, node.Hook, "", map[string]string{})
	if err != nil {
		return err
	}

	_, err = config.NodeAdd(cfg, p.ID, dele.ID, pno2.ID, node.Delay, "", map[string]string{dele.ID: delayi.ID})
	if err != nil {
		return err
	}

	return nil
}

func useradd(cfg database.Config, email, password string) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

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
