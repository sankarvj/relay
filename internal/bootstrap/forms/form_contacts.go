package forms

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func ContactFields(ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         schema.SeedFieldFNameKey,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle, entity.MetaKeyUnique: "true"},
	}

	emailField := entity.Field{
		Key:         "uuid-00-email",
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	mobileField := entity.Field{
		Key:         "uuid-00-mobile-numbers",
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	avatarField := entity.Field{
		Key:         "uuid-00-avatar",
		Name:        "avatar",
		DisplayName: "Avatar",
		DataType:    entity.TypeString,
		DomType:     entity.DomImage,
		Who:         entity.WhoAvatar,
	}

	npsField := entity.Field{
		Key:         schema.SeedFieldNPSKey,
		Name:        "nps_score",
		DisplayName: "NPS Score",
		DataType:    entity.TypeNumber,
		DomType:     entity.DomText,
	}

	lfStageField := entity.Field{
		Key:         "uuid-00-lf-stage",
		Name:        "lifecycle_stage",
		DisplayName: "Lifecycle Stage",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "1",
				DisplayValue: "Lead",
			},
			{
				ID:           "2",
				DisplayValue: "Contact",
			},
		},
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	ownerField := entity.Field{
		Key:         "uuid-00-owner",
		Name:        "owner",
		DisplayName: "Owner",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, avatarField, ownerField}
}

func ContactVals(name, email string) map[string]interface{} {
	contactVals := map[string]interface{}{
		schema.SeedFieldFNameKey: name,
		"uuid-00-email":          email,
		"uuid-00-mobile-numbers": []interface{}{"9944293499", "9940209164"},
		schema.SeedFieldNPSKey:   100,
		"uuid-00-lf-stage":       []interface{}{"1"},
		"uuid-00-owner":          []interface{}{},
	}
	return contactVals
}
