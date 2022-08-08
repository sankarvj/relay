package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func ServiceCategory() []entity.Field {
	serviceCategoryNameFieldID := uuid.New().String()
	serviceCategoryNameField := entity.Field{
		Key:         serviceCategoryNameFieldID,
		Name:        "service_category",
		DisplayName: "Service Category",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	imageFieldID := uuid.New().String()
	imageField := entity.Field{
		Key:         imageFieldID,
		Name:        "image",
		DisplayName: "Image",
		DataType:    entity.TypeString,
		DomType:     entity.DomImage,
		Who:         entity.WhoImage,
	}

	return []entity.Field{serviceCategoryNameField, imageField}
}

func ServiceFields(serviceCatagoryEntityID, serviceCatagoryDisplayKey string) []entity.Field {
	serviceNameFieldID := uuid.New().String()
	serviceNameField := entity.Field{
		Key:         serviceNameFieldID,
		Name:        "service_name",
		DisplayName: "Service Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	serviceCategoryFieldID := uuid.New().String()
	serviceCategoryField := entity.Field{
		Key:         serviceCategoryFieldID,
		Name:        "catagory",
		DisplayName: "Catagory",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       serviceCatagoryEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: serviceCatagoryDisplayKey},
		Who:         entity.WhoAssetCategory,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{serviceNameField, serviceCategoryField}
}

func ServiceVals(serviceNamekey, catKey, name string, cat []interface{}) map[string]interface{} {
	serviceVals := map[string]interface{}{
		serviceNamekey: name,
		catKey:         cat,
	}
	return serviceVals
}
