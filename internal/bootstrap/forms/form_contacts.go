package forms

import (
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func ContactFields(ownerEntityID string, ownerEntityKey string) []entity.Field {
	firstNameFieldID := uuid.New().String()
	firstNameField := entity.Field{
		Key:         firstNameFieldID,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle, entity.MetaKeyUnique: "true"},
	}

	lastNameFieldID := uuid.New().String()
	lastNameField := entity.Field{
		Key:         lastNameFieldID,
		Name:        "last_name",
		DisplayName: "Last Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoEmail,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	mobileFieldID := uuid.New().String()
	mobileField := entity.Field{
		Key:         mobileFieldID,
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	avatarFieldID := uuid.New().String()
	avatarField := entity.Field{
		Key:         avatarFieldID,
		Name:        "avatar",
		DisplayName: "Avatar",
		DataType:    entity.TypeString,
		DomType:     entity.DomImage,
		Who:         entity.WhoAvatar,
	}

	npsFieldID := uuid.New().String()
	npsField := entity.Field{
		Key:         npsFieldID,
		Name:        "nps_score",
		DisplayName: "NPS Score",
		DataType:    entity.TypeNumber,
		DomType:     entity.DomText,
	}

	lfStageFieldID := uuid.New().String()
	lfStageField := entity.Field{
		Key:         lfStageFieldID,
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
				DisplayValue: "Sales Qualified Lead",
			},
			{
				ID:           "3",
				DisplayValue: "Customer",
			},
			{
				ID:           "4",
				DisplayValue: "Other",
			},
		},
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
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

	return []entity.Field{firstNameField, lastNameField, emailField, mobileField, npsField, lfStageField, avatarField, ownerField}
}

func ContactVals(contactEntity entity.Entity, firstName, email string) map[string]interface{} {

	namedVals := map[string]interface{}{
		"first_name":      firstName,
		"email":           email,
		"mobile_numbers":  []interface{}{randomdata.PhoneNumber(), randomdata.PhoneNumber()},
		"nps_score":       randomdata.Number(100),
		"lifecycle_stage": []interface{}{"1"},
		"owner":           []interface{}{},
		"avatar":          fmt.Sprintf("https://avatars.dicebear.com/api/avataaars/%s.svg", firstName),
	}

	return keyMap(contactEntity.NamedKeys(), namedVals)
}
