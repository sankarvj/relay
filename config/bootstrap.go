package config

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func EntityAdd(cfg database.Config, name string, cat int, fields []entity.Field) (entity.Entity, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return entity.Entity{}, err
	}
	defer db.Close()

	ctx := context.Background()
	ne := entity.NewEntity{
		AccountID: schema.SeedAccountID,
		TeamID:    schema.SeedTeamID,
		Category:  cat,
		Name:      name,
		Fields:    fields,
	}

	e, err := entity.Create(ctx, db, ne, time.Now())
	if err != nil {
		return entity.Entity{}, err
	}

	fmt.Println("Entity created with id:", e.ID)
	return e, nil
}

func ItemAdd(cfg database.Config, entityID string, fields map[string]interface{}) (item.Item, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return item.Item{}, err
	}
	defer db.Close()

	ctx := context.Background()
	ni := item.NewItem{
		AccountID: schema.SeedAccountID,
		EntityID:  entityID,
		Fields:    fields,
	}

	i, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	fmt.Println("Item created with id:", i.ID)
	return i, nil
}

func FlowAdd(cfg database.Config, entityID string, name string, typ, condition int) (flow.Flow, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return flow.Flow{}, err
	}
	defer db.Close()

	ctx := context.Background()
	nf := flow.NewFlow{
		AccountID:  schema.SeedAccountID,
		EntityID:   entityID,
		Type:       typ,
		Condition:  condition,
		Expression: `{{` + entityID + `.uuid-00-fname}} eq {Vijay} && {{` + entityID + `.uuid-00-nps-score}} gt {98}`,
		Name:       name,
	}

	f, err := flow.Create(ctx, db, nf, time.Now())
	if err != nil {
		return flow.Flow{}, err
	}

	fmt.Println("Flow created with id:", f.ID)
	return f, nil
}

func NodeAdd(cfg database.Config, flowID, actorID, pnodeID string, typ int, exp string, actuals map[string]string) (node.Node, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return node.Node{}, err
	}
	defer db.Close()

	ctx := context.Background()
	nn := node.NewNode{
		AccountID:    schema.SeedAccountID,
		FlowID:       flowID,
		ActorID:      actorID,
		ParentNodeID: pnodeID,
		Type:         typ,
		Expression:   exp,
		Actuals:      actuals,
	}

	n, err := node.Create(ctx, db, nn, time.Now())
	if err != nil {
		return node.Node{}, err
	}

	fmt.Println("Node created with id:", n.ID)
	return n, nil
}

func StatusFields() []entity.Field {
	nameField := entity.Field{
		Key:      "uuid-00-name",
		Name:     "name",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	colorField := entity.Field{
		Key:      "uuid-00-color",
		Name:     "color",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	return []entity.Field{nameField, colorField}
}

func StatusVals(name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		"uuid-00-name":  name,
		"uuid-00-color": color,
	}
	return statusVals
}

func ContactFields(statusEntityID string) []entity.Field {
	nameField := entity.Field{
		Key:      "uuid-00-fname",
		Name:     "First Name",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	emailField := entity.Field{
		Key:      "uuid-00-email",
		Name:     "Email",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	mobileField := entity.Field{
		Key:      "uuid-00-mobile-numbers",
		Name:     "Mobile Numbers",
		DataType: entity.TypeList,
		DomType:  entity.DomMultiSelect,
		Field: &entity.Field{
			DataType: entity.TypeString,
		},
	}

	npsField := entity.Field{
		Key:      "uuid-00-nps-score",
		Name:     "NPS Score",
		DataType: entity.TypeNumber,
	}

	lfStageField := entity.Field{
		Key:      "uuid-00-lf-stage",
		Name:     "Lifecycle Stage",
		DomType:  entity.DomSelect,
		DataType: entity.TypeString,
		Choices:  []string{"lead", "contact", "won"},
	}

	statusField := entity.Field{
		Key:      "uuid-00-status",
		Name:     "Status",
		DomType:  entity.DomText,
		DataType: entity.TypeReference,
		RefID:    statusEntityID,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, statusField}
}

func ContactVals(name, email, statusID string) map[string]interface{} {
	contactVals := map[string]interface{}{
		"uuid-00-fname":          name,
		"uuid-00-email":          email,
		"uuid-00-mobile-numbers": []string{"9944293499", "9940209164"},
		"uuid-00-nps-score":      100,
		"uuid-00-lf-stage":       "lead",
		"uuid-00-status":         []string{statusID},
	}
	return contactVals
}

func TaskFields(contactEntityID string) []entity.Field {
	descField := entity.Field{
		Key:      "uuid-00-desc",
		Name:     "desc",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	contactField := entity.Field{
		Key:      "uuid-00-contact",
		Name:     "Contact",
		DomType:  entity.DomText,
		DataType: entity.TypeReference,
		RefID:    contactEntityID,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, contactField}
}

func TaskVals(desc, contactID string) map[string]interface{} {
	taskVals := map[string]interface{}{
		"uuid-00-desc":    desc,
		"uuid-00-contact": []string{contactID},
	}
	return taskVals
}

func DealFields(contactEntityID string) []entity.Field {
	dealName := entity.Field{
		Key:      "uuid-00-deal-name",
		Name:     "Deal Name",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	dealAmount := entity.Field{
		Key:      "uuid-00-deal-amount",
		Name:     "Deal Amount",
		DomType:  entity.DomText,
		DataType: entity.TypeNumber,
	}

	contactsField := entity.Field{
		Key:      "uuid-00-contacts",
		Name:     "Contacts",
		DomType:  entity.DomText,
		DataType: entity.TypeReference,
		RefID:    contactEntityID,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{dealName, dealAmount, contactsField}
}

func DealVals(name string, amount int, contactID1, contactID2 string) map[string]interface{} {
	dealVals := map[string]interface{}{
		"uuid-00-deal-name":   name,
		"uuid-00-deal-amount": amount,
		"uuid-00-contacts":    []string{contactID1, contactID2},
	}
	return dealVals
}

func EmailFields() []entity.Field {
	domain := entity.Field{
		Key:      "uuid-00-domain",
		Name:     "domain",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "sandbox3ab4868d173f4391805389718914b89c.mailgun.org",
		Meta: map[string]string{
			"config": "true",
		},
	}

	apiKey := entity.Field{
		Key:      "uuid-00-api-key",
		Name:     "api_key",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35",
		Meta: map[string]string{
			"config": "true",
		},
	}

	sender := entity.Field{
		Key:      "uuid-00-sender",
		Name:     "sender",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "vijayasankar.jothi@wayplot.com",
	}

	to := entity.Field{
		Key:      "uuid-00-to",
		Name:     "to",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "",
	}

	cc := entity.Field{
		Key:      "uuid-00-cc",
		Name:     "cc",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "",
	}

	subject := entity.Field{
		Key:      "uuid-00-subject",
		Name:     "subject",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "",
	}

	body := entity.Field{
		Key:      "uuid-00-body",
		Name:     "body",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "",
	}

	return []entity.Field{domain, apiKey, sender, to, cc, subject, body}
}

func EmailVals(contactEntityID string) map[string]interface{} {
	emailValues := map[string]interface{}{
		"uuid-00-domain":  "sandbox3ab4868d173f4391805389718914b89c.mailgun.org",
		"uuid-00-api-key": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35",
		"uuid-00-sender":  "vijayasankar.jothi@wayplot.com",
		"uuid-00-to":      `{{` + contactEntityID + `.uuid-00-email}}`,
		"uuid-00-cc":      "vijayasankarmobile@gmail.com",
		"uuid-00-subject": `This mail is sent you to tell that your NPS scrore is {{` + contactEntityID + `.uuid-00-nps-score}}. We are very proud of you! `,
		"uuid-00-body":    `Hello {{` + contactEntityID + `.uuid-00-fname}}`,
	}
	return emailValues
}

func APIFields() []entity.Field {
	path := entity.Field{
		Key:      "uuid-00-path",
		Name:     "path",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "/actuator/info",
	}

	host := entity.Field{
		Key:      "uuid-00-host",
		Name:     "host",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "https://stage.freshcontacts.io",
	}

	method := entity.Field{
		Key:      "uuid-00-method",
		Name:     "method",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "GET",
	}

	headers := entity.Field{
		Key:      "uuid-00-headers",
		Name:     "headers",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}",
	}

	return []entity.Field{path, host, method, headers}
}

func DelayVals() map[string]interface{} {
	delayVals := map[string]interface{}{
		"uuid-00-delay-by": "2",
		"uuid-00-repeat":   "true",
	}
	return delayVals
}

func DelayFields() []entity.Field {
	delay_by := entity.Field{
		Key:      "uuid-00-delay-by",
		Name:     "delay_by",
		DomType:  entity.DomText,
		DataType: entity.TypeDataTime,
	}

	repeat := entity.Field{
		Key:      "uuid-00-repeat",
		Name:     "repeat",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "true",
	}

	return []entity.Field{delay_by, repeat}
}
