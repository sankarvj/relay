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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
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

	if err := schema.SeedUsers(db); err != nil {
		return err
	}

	ctx := context.Background()
	cuser, err := user.RetrieveUser(ctx, db, schema.SeedUserID1)
	if err != nil {
		return err
	}
	accountID := schema.SeedAccountID
	teamID := schema.SeedTeamID
	nc := account.NewAccount{
		Name:   "Wayplot",
		Domain: "wayplot.com"}
	err = account.AccountBootstrap(ctx, db, cuser, accountID, teamID, nc, time.Now())
	if err != nil {
		log.Println("!!!! TODO: Should Implement Roll Back Option Here.")
		return err
	}

	fmt.Println("Seed data complete")
	return nil
}

func crmadd(cfg database.Config) error {
	accountID := schema.SeedAccountID
	teamID := schema.SeedTeamID

	fmt.Printf("CRM Bootstrap request received for accountID %s and teamID %s \n", accountID, teamID)

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityOwner)
	if err != nil {
		return err
	}

	//add entity - status
	se, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Status", entity.CategoryChildUnit, bootstrap.StatusFields())
	if err != nil {
		return err
	}
	// add status item - open
	st1, err := bootstrap.ItemAdd(ctx, db, accountID, se.ID, uuid.New().String(), bootstrap.StatusVals("Open", "#fb667e"))
	if err != nil {
		return err
	}
	// add status item - closed
	st2, err := bootstrap.ItemAdd(ctx, db, accountID, se.ID, uuid.New().String(), bootstrap.StatusVals("Closed", "#66fb99"))
	if err != nil {
		return err
	}
	// add status item - overdue
	st3, err := bootstrap.ItemAdd(ctx, db, accountID, se.ID, uuid.New().String(), bootstrap.StatusVals("OverDue", "#66fb99"))
	if err != nil {
		return err
	}

	//add entity - contacts
	ce, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Contacts", entity.CategoryData, bootstrap.ContactFields(se.ID, ownerEntity.ID))
	if err != nil {
		return err
	}
	// add contact item - vijay (straight)
	con1, err := bootstrap.ItemAdd(ctx, db, accountID, ce.ID, uuid.New().String(), bootstrap.ContactVals("Vijay", "vijayasankarmail@gmail.com", st1.ID))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	con2, err := bootstrap.ItemAdd(ctx, db, accountID, ce.ID, uuid.New().String(), bootstrap.ContactVals("Senthil", "vijayasankarmail@gmail.com", st2.ID))
	if err != nil {
		return err
	}

	//add entity - companies
	come, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Companies", entity.CategoryData, bootstrap.CompanyFields())
	if err != nil {
		return err
	}
	com1, err := bootstrap.ItemAdd(ctx, db, accountID, come.ID, uuid.New().String(), bootstrap.CompanyVals("Zoho", "zoho.com"))
	if err != nil {
		return err
	}

	//add entity - task
	te, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Tasks", entity.CategoryData, bootstrap.TaskFields(ce.ID, se.ID, st1.ID, st2.ID, st3.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = bootstrap.ItemAdd(ctx, db, accountID, te.ID, uuid.New().String(), bootstrap.TaskVals("make cake", con1.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = bootstrap.ItemAdd(ctx, db, accountID, te.ID, uuid.New().String(), bootstrap.TaskVals("make call", con1.ID))
	if err != nil {
		return err
	}

	ei, err := entity.SaveEmailIntegration(ctx, accountID, schema.SeedUserID1, "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "vijayasankar.jothi@wayplot.com", db)
	if err != nil {
		return err
	}

	fields, _ := ce.Fields()
	namedKeysMap := entity.NamedKeysMap(fields)
	to := fmt.Sprintf("{{%s.%s}}", ce.ID, namedKeysMap["email"])
	cc := "vijayasankarmobile@gmail.com"
	subject := fmt.Sprintf("This mail is sent you to tell that your NPS scrore is {{%s.%s}}. We are very proud of you!", ce.ID, namedKeysMap["nps_score"])
	body := fmt.Sprintf("Hello {{%s.%s}}", ce.ID, namedKeysMap["email"])
	emg, err := entity.SaveEmailTemplate(ctx, accountID, ei.ID, to, cc, "", subject, body, db)
	if err != nil {
		return err
	}

	//add entity - api-hook
	we, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "WebHook", entity.CategoryAPI, bootstrap.APIFields())
	if err != nil {
		return err
	}

	//add entity - delay
	dele, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Delay Timer", entity.CategoryDelay, bootstrap.DelayFields())
	if err != nil {
		return err
	}
	// add delay item
	delayi, err := bootstrap.ItemAdd(ctx, db, accountID, dele.ID, uuid.New().String(), bootstrap.DelayVals())
	if err != nil {
		return err
	}

	pID, nID, err := addPipelines(ctx, db, accountID, ce.ID, ei.ID, we.ID, dele.ID, emg.ID, delayi.ID)
	if err != nil {
		return err
	}

	//add entity - deal
	de, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Deals", entity.CategoryData, bootstrap.DealFields(ce.ID, pID))
	if err != nil {
		return err
	}
	// add deal item with contacts - vijay & senthil (reverse) & pipeline stage
	deal1, err := bootstrap.ItemAdd(ctx, db, accountID, de.ID, uuid.New().String(), bootstrap.DealVals("Big Deal", 1000, con1.ID, con2.ID, nID))
	if err != nil {
		return err
	}

	//add entity - tickets
	tice, err := bootstrap.EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), "", "Tickets", entity.CategoryData, bootstrap.TicketFields(se.ID))
	if err != nil {
		return err
	}

	ticket1, err := bootstrap.ItemAdd(ctx, db, accountID, tice.ID, uuid.New().String(), bootstrap.TicketVals("My Laptop Is Not Working", st1.ID))
	if err != nil {
		return err
	}

	//contact company association
	associationID, err := bootstrap.AssociationAdd(ctx, db, accountID, ce.ID, come.ID)
	if err != nil {
		return err
	}
	err = bootstrap.ConnectionAdd(ctx, db, accountID, associationID, con1.ID, com1.ID)
	if err != nil {
		return err
	}

	//ticket deal association
	tdaID, err := bootstrap.AssociationAdd(ctx, db, accountID, tice.ID, de.ID)
	if err != nil {
		return err
	}
	err = bootstrap.ConnectionAdd(ctx, db, accountID, tdaID, ticket1.ID, deal1.ID)
	if err != nil {
		return err
	}

	//ticket contact association
	tcaID, err := bootstrap.AssociationAdd(ctx, db, accountID, tice.ID, ce.ID)
	if err != nil {
		return err
	}
	err = bootstrap.ConnectionAdd(ctx, db, accountID, tcaID, ticket1.ID, con1.ID)
	if err != nil {
		return err
	}

	// //add workflows
	// f, err := config.FlowAdd(cfg, "00000000-0000-0000-0000-000000000017", ce.ID, "The Workflow", flow.FlowModeWorkFlow, flow.FlowConditionEntry)
	// if err != nil {
	// 	return err
	// }

	// //test node push test case - TestCreateItemRuleRunner
	// no1, err := config.NodeAdd(cfg, "00000000-0000-0000-0000-000000000018", f.ID, te.ID, node.Root, node.Push, "", map[string]string{})
	// if err != nil {
	// 	return err
	// }

	// no2, err := config.NodeAdd(cfg, "00000000-0000-0000-0000-000000000019", f.ID, "00000000-0000-0000-0000-000000000000", no1.ID, node.Decision, "{Vijay} eq {Vijay}", map[string]string{})
	// if err != nil {
	// 	return err
	// }

	// no3, err := config.NodeAdd(cfg, "00000000-0000-0000-0000-000000000020", f.ID, me.ID, no2.ID, node.Email, "{{xyz.result}} eq {true}", map[string]string{me.ID: emg.ID})
	// if err != nil {
	// 	return err
	// }

	// _, err = config.NodeAdd(cfg, "00000000-0000-0000-0000-000000000021", f.ID, we.ID, no2.ID, node.Hook, "{{xyz.result}} eq {false}", map[string]string{})
	// if err != nil {
	// 	return err
	// }

	// _, err = config.NodeAdd(cfg, "00000000-0000-0000-0000-000000000022", f.ID, dele.ID, no3.ID, node.Delay, "", map[string]string{dele.ID: delayi.ID})
	// if err != nil {
	// 	return err
	// }

	return nil
}

func addPipelines(ctx context.Context, db *sqlx.DB, accountID, contactEntityID, mailEntityID, webhookEntityID, delayEntityID, mailItemID, delayItemID string) (string, string, error) {
	//add pipelines
	p, err := bootstrap.FlowAdd(ctx, db, accountID, uuid.New().String(), contactEntityID, "Sales Pipeline", flow.FlowModePipeLine, flow.FlowConditionEntry)
	if err != nil {
		return "", "", err
	}

	pno1, err := bootstrap.NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, "00000000-0000-0000-0000-000000000000", node.Root, "opportunity", node.Stage, "", map[string]string{})
	if err != nil {
		return "", "", err
	}

	pno2, err := bootstrap.NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, "00000000-0000-0000-0000-000000000000", pno1.ID, "Deal Won", node.Stage, "{Vijay} eq {Vijay}", map[string]string{})
	if err != nil {
		return "", "", err
	}

	_, err = bootstrap.NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, mailEntityID, pno1.ID, "", node.Email, "", map[string]string{mailEntityID: mailItemID})
	if err != nil {
		return "", "", err
	}

	_, err = bootstrap.NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, webhookEntityID, pno1.ID, "", node.Hook, "", map[string]string{})
	if err != nil {
		return "", "", err
	}

	_, err = bootstrap.NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, delayEntityID, pno2.ID, "", node.Delay, "", map[string]string{delayEntityID: delayItemID})
	if err != nil {
		return "", "", err
	}
	return p.ID, pno1.ID, nil
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
