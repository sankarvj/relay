package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func AssetCategory() []entity.Field {
	assetCategoryNameFieldID := uuid.New().String()
	assetCategoryNameField := entity.Field{
		Key:         assetCategoryNameFieldID,
		Name:        "asset_category",
		DisplayName: "Asset Category",
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

	return []entity.Field{assetCategoryNameField, imageField}
}

func AssetFields(assetCatagoryEntityID, assetCatagoryDisplayKey string) []entity.Field {
	assetNameFieldID := uuid.New().String()
	assetNameField := entity.Field{
		Key:         assetNameFieldID,
		Name:        "asset_name",
		DisplayName: "Asset Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	assetCategoryFieldID := uuid.New().String()
	assetCategoryField := entity.Field{
		Key:         assetCategoryFieldID,
		Name:        "catagory",
		DisplayName: "Catagory",
		DomType:     entity.DomAutoSelect,
		DataType:    entity.TypeReference,
		RefID:       assetCatagoryEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: assetCatagoryDisplayKey},
		Who:         entity.WhoAssetCategory,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{assetNameField, assetCategoryField}
}

func AssetVals(assetNamekey, catKey, name string, cat []interface{}) map[string]interface{} {
	assetVals := map[string]interface{}{
		assetNamekey: name,
		catKey:       cat,
	}
	return assetVals
}
