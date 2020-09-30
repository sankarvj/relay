package config

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
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
