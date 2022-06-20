package em

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)
	b.OwnerEntity.Fields()
	ownerKey := b.OwnerEntity.NamedKeys()["email"]
	fmt.Println("\tEM:BOOT Owner Entity Retrived")

	// add entity - roles
	rolesEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityRoles, "Roles", entity.CategoryChildUnit, entity.StateTeamLevel, RoleFields())
	if err != nil {
		return err
	}
	roleEntityKey := rolesEntity.NamedKeys()["role"]
	fmt.Println("\tEM:BOOT Role Entity Created")

	// add role item - Intern
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Intern"), nil)
	if err != nil {
		return err
	}
	// add role item - Developer
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Developer"), nil)
	if err != nil {
		return err
	}
	// add role item - QA Engineer
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "QA Engineer"), nil)
	if err != nil {
		return err
	}
	// add role item - Manager
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Manager"), nil)
	if err != nil {
		return err
	}
	// add role item - Sales Rep
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Sales Rep"), nil)
	if err != nil {
		return err
	}
	// add role item - Support Engineer
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Support Engineer"), nil)
	if err != nil {
		return err
	}
	// add role item - Director
	_, err = b.ItemAdd(ctx, rolesEntity.ID, uuid.New().String(), b.UserID, RoleVals(roleEntityKey, "Director"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Roles Created")

	// add entity - employees
	employeeEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmployee, "Employees", entity.CategoryData, entity.StateTeamLevel, EmployeeFields(b.OwnerEntity.ID, ownerKey, rolesEntity.ID, roleEntityKey))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Employee Entity Created")
	// add entity - asset-catagory
	assetCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetCatagory, "Assets Category", entity.CategorySubData, entity.StateTeamLevel, forms.AssetCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Catagory Entity Created")
	// add entity - assets
	assetCatagoryTitleField := entity.TitleField(assetCatagoryEntity.FieldsIgnoreError())
	assetEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssets, "Assets", entity.CategoryChildUnit, entity.StateTeamLevel, forms.AssetFields(assetCatagoryEntity.ID, assetCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Assets Entity Created")
	assetsFieldsNamedKeysMap := assetEntity.NamedKeys()
	// add asset item - MAC
	_, err = b.ItemAdd(ctx, assetEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "Mac 24 Inch", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add asset item - Mobile
	_, err = b.ItemAdd(ctx, assetEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "iPhone 16", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add asset item - Mouse
	_, err = b.ItemAdd(ctx, assetEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "Mouse", []interface{}{}), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Assets Created")

	// add entity - asset status
	assetStatusEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityStatus, "Asset Status", entity.CategoryChildUnit, entity.StateTeamLevel, AssetStatusFields())
	if err != nil {
		return err
	}
	statusFieldsNamedKeysMap := assetStatusEntity.NamedKeys()
	// add status item - Request-received
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Request Received", "#fb667e"), nil)
	if err != nil {
		return err
	}
	// add status item - In-progress
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "In Progress", "#66fb99"), nil)
	if err != nil {
		return err
	}
	// add status item - Approved
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Approved", "#66fb99"), nil)
	if err != nil {
		return err
	}
	// add status item - Delivered
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Delivered", "#66fb99"), nil)
	if err != nil {
		return err
	}
	// add status item - Returned
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Returned", "#66fb99"), nil)
	if err != nil {
		return err
	}
	// add status item - Trashed
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Trashed", "#66fb99"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Status Entity Created")

	// add entity - asset request
	assetStatusTitleField := entity.TitleField(assetStatusEntity.FieldsIgnoreError())
	assetTitleField := entity.TitleField(assetEntity.FieldsIgnoreError())
	assetRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetRequest, "Asset Request", entity.CategorySubData, entity.StateTeamLevel, AssetRequestFields(assetEntity.ID, assetTitleField.Key, assetStatusEntity.ID, assetStatusTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Request Entity Created")

	// add entity - service-catagory
	serviceCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceCatagory, "Service Category", entity.CategorySubData, entity.StateTeamLevel, forms.ServiceCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Catagory Entity Created")
	// add entity - services
	serviceCatagoryTitleField := entity.TitleField(serviceCatagoryEntity.FieldsIgnoreError())
	serviceEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServices, "Services", entity.CategoryChildUnit, entity.StateTeamLevel, forms.ServiceFields(serviceCatagoryEntity.ID, serviceCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Services Entity Created")
	servicesFieldsNamedKeysMap := serviceEntity.NamedKeys()
	// add service item - Git Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Git access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Mail Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Mail access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Slack Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Slack access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Infra Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Infra access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Marketing Tools Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Marketing tools access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Earnings Access
	_, err = b.ItemAdd(ctx, serviceEntity.ID, uuid.New().String(), b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Earning sheet access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Services Created")
	// add entity - asset request
	serviceTitleField := entity.TitleField(serviceEntity.FieldsIgnoreError())
	serviceRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceRequest, "Service Request", entity.CategorySubData, entity.StateTeamLevel, ServiceRequestFields(serviceEntity.ID, serviceTitleField.Key, assetStatusEntity.ID, assetStatusTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Request Entity Created")

	err = AddAssociations(ctx, b, employeeEntity, assetRequestEntity, serviceRequestEntity)
	if err != nil {
		return err
	}
	err = AddSamples(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	return nil
}

func AddAssociations(ctx context.Context, b *base.Base, employeeEid, assetRequestEid, serviceRequestEid entity.Entity) error {

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	//employee asset-request association
	_, err = b.AssociationAdd(ctx, employeeEid.ID, assetRequestEid.ID)
	if err != nil {
		return err
	}

	//employee service-request association
	_, err = b.AssociationAdd(ctx, employeeEid.ID, serviceRequestEid.ID)
	if err != nil {
		return err
	}

	//employee stream association
	_, err = b.AssociationAdd(ctx, employeeEid.ID, streamEntity.ID)
	if err != nil {
		return err
	}

	//asset-request stream association
	_, err = b.AssociationAdd(ctx, assetRequestEid.ID, streamEntity.ID)
	if err != nil {
		return err
	}

	//service-request stream association
	_, err = b.AssociationAdd(ctx, serviceRequestEid.ID, streamEntity.ID)
	if err != nil {
		return err
	}

	return nil
}

func AddSamples(ctx context.Context, b *base.Base) error {
	return nil
}
