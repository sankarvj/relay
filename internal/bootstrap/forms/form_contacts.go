package forms

import (
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func ContactFields(ownerEntityID, ownerEntityKey string, companyEntityID, companyEntityKey string, leadStatusEntityID, leadStatusEntityKey string) []entity.Field {
	fields := make([]entity.Field, 0)

	firstNameFieldID := uuid.New().String()
	firstNameField := entity.Field{
		Key:         firstNameFieldID,
		Name:        "first_name",
		DisplayName: "First name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}
	fields = append(fields, firstNameField)

	lastNameFieldID := uuid.New().String()
	lastNameField := entity.Field{
		Key:         lastNameFieldID,
		Name:        "last_name",
		DisplayName: "Last name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}
	fields = append(fields, lastNameField)

	jobTitleFieldID := uuid.New().String()
	jobTitleField := entity.Field{
		Key:         jobTitleFieldID,
		Name:        "job_title",
		DisplayName: "Job title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{},
	}
	fields = append(fields, jobTitleField)

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoEmail,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle, entity.MetaKeyUnique: "true"},
	}
	fields = append(fields, emailField)

	mobileFieldID := uuid.New().String()
	mobileField := entity.Field{
		Key:         mobileFieldID,
		Name:        "mobile_numbers",
		DisplayName: "Mobile numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}
	fields = append(fields, mobileField)

	avatarFieldID := uuid.New().String()
	avatarField := entity.Field{
		Key:         avatarFieldID,
		Name:        "avatar",
		DisplayName: "Avatar",
		DataType:    entity.TypeString,
		DomType:     entity.DomImage,
		Who:         entity.WhoAvatar,
	}
	fields = append(fields, avatarField)

	npsFieldID := uuid.New().String()
	npsField := entity.Field{
		Key:         npsFieldID,
		Name:        "nps_score",
		DisplayName: "NPS score",
		DataType:    entity.TypeNumber,
		DomType:     entity.DomText,
	}
	fields = append(fields, npsField)

	lfStageFieldID := uuid.New().String()
	lfStageField := entity.Field{
		Key:         lfStageFieldID,
		Name:        "lifecycle_stage",
		DisplayName: "Lifecycle stage",
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
				DisplayValue: "InActive",
			},
			{
				ID:           "5",
				DisplayValue: "Other",
			},
		},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}
	fields = append(fields, lfStageField)

	if leadStatusEntityID != "" {
		leadStatusFieldID := uuid.New().String()
		leadStatusField := entity.Field{
			Key:         leadStatusFieldID,
			Name:        "lead_status",
			DisplayName: "Lead Status",
			DomType:     entity.DomSelect,
			DataType:    entity.TypeReference,
			RefID:       leadStatusEntityID,
			RefType:     entity.RefTypeStraight,
			Meta:        map[string]string{entity.MetaKeyDisplayGex: leadStatusEntityKey},
			Field: &entity.Field{
				DataType: entity.TypeString,
				Key:      "id",
				Value:    "--",
			},
		}
		fields = append(fields, leadStatusField)
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
		Name:        "owner",
		DisplayName: "Contact owner",
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
	fields = append(fields, ownerField)

	associatedCompaniesFieldID := uuid.New().String()
	associatedCompaniesField := entity.Field{
		Key:         associatedCompaniesFieldID,
		Name:        "associated_companies",
		DisplayName: "Associated companies",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}
	fields = append(fields, associatedCompaniesField)

	createdByUserFieldID := uuid.New().String()
	createdByUserField := entity.Field{
		Key:         createdByUserFieldID,
		Name:        "created_by_user",
		DisplayName: "Created by user",
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
	fields = append(fields, createdByUserField)

	becameACustomerDateFieldID := uuid.New().String()
	becameACustomerDateField := entity.Field{
		Key:         becameACustomerDateFieldID,
		Name:        "became_a_customer_date",
		DisplayName: "Became a customer date",
		DomType:     entity.DomText,
		DataType:    entity.TypeDate,
	}
	fields = append(fields, becameACustomerDateField)

	lostCustomerDateFieldID := uuid.New().String()
	lostCustomerDateField := entity.Field{
		Key:         lostCustomerDateFieldID,
		Name:        "lost_customer_on",
		DisplayName: "Lost customer on",
		DomType:     entity.DomText,
		DataType:    entity.TypeDate,
	}
	fields = append(fields, lostCustomerDateField)

	totalRevenueFieldID := uuid.New().String()
	totalRevenueField := entity.Field{
		Key:         totalRevenueFieldID,
		Name:        "total_revenue",
		DisplayName: "Total revenue",
		DataType:    entity.TypeNumber,
		DomType:     entity.DomText,
		Meta:        map[string]string{entity.MetaKeyCalc: entity.MetaCalcSum},
	}
	fields = append(fields, totalRevenueField)

	timeZoneFieldID := uuid.New().String()
	timeZoneField := entity.Field{
		Key:         timeZoneFieldID,
		Name:        "time_zone",
		DisplayName: "Time zone",
		DataType:    entity.TypeString,
		DomType:     entity.DomText,
	}
	fields = append(fields, timeZoneField)

	websiteURLFieldID := uuid.New().String()
	websiteURLField := entity.Field{
		Key:         websiteURLFieldID,
		Name:        "website_url",
		DisplayName: "Website URL",
		DataType:    entity.TypeString,
		DomType:     entity.DomText,
	}
	fields = append(fields, websiteURLField)

	twitterUserNameFieldID := uuid.New().String()
	twitterUserNameField := entity.Field{
		Key:         twitterUserNameFieldID,
		Name:        "twitter_username",
		DisplayName: "Twitter username",
		DataType:    entity.TypeString,
		DomType:     entity.DomText,
	}
	fields = append(fields, twitterUserNameField)

	tagsFieldID := uuid.New().String()
	tagsField := entity.Field{
		Key:         tagsFieldID,
		Name:        "tags",
		DisplayName: "Tags",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Meta:        map[string]string{entity.MetaKeyCalc: entity.MetaCalcAggr},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}
	fields = append(fields, tagsField)

	return fields
}

func ContactVals(contactEntity entity.Entity, firstName, lastName, email, leadStatusItemID string) map[string]interface{} {

	namedVals := map[string]interface{}{
		"first_name":             firstName,
		"last_name":              lastName,
		"email":                  email,
		"mobile_numbers":         []interface{}{randomdata.PhoneNumber(), randomdata.PhoneNumber()},
		"nps_score":              randomdata.Number(100),
		"lifecycle_stage":        []interface{}{util.ConvertIntToStr(randomdata.Number(1, 5))},
		"owner":                  []interface{}{},
		"avatar":                 fmt.Sprintf("https://avatars.dicebear.com/api/pixel-art/%s.svg", firstName),
		"lead_status":            []interface{}{leadStatusItemID},
		"became_a_customer_date": util.FormatTimeGo(time.Now()),
	}

	return keyMap(contactEntity.NameKeyMapWrapper(), namedVals)
}
