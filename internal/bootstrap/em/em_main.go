package em

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)
	b.OwnerEntity.Fields()
	ownerKey := b.OwnerEntity.NamedKeys()["email"]
	fmt.Println("\tEM:BOOT Owner Entity Retrived")

	itemIDMap := make(map[string]string, 0)

	// add entity - roles
	rolesEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityRoles, "Roles", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, RoleFields())
	if err != nil {
		return err
	}
	roleEntityKey := rolesEntity.NamedKeys()["role"]
	fmt.Println("\tEM:BOOT Role Entity Created")

	// add role item - Intern
	itemIDMap["intern"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["intern"], b.UserID, RoleVals(roleEntityKey, "Intern"), nil)
	if err != nil {
		return err
	}
	// add role item - Developer
	itemIDMap["developer"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["developer"], b.UserID, RoleVals(roleEntityKey, "Developer"), nil)
	if err != nil {
		return err
	}

	// add role item - designer
	itemIDMap["designer"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["designer"], b.UserID, RoleVals(roleEntityKey, "Designer"), nil)
	if err != nil {
		return err
	}

	// add role item - QA Engineer
	itemIDMap["qa_engineer"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["qa_engineer"], b.UserID, RoleVals(roleEntityKey, "QA Engineer"), nil)
	if err != nil {
		return err
	}
	// add role item - Manager
	itemIDMap["manager"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["manager"], b.UserID, RoleVals(roleEntityKey, "Manager"), nil)
	if err != nil {
		return err
	}
	// add role item - Sales Rep
	itemIDMap["sales_rep"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["sales_rep"], b.UserID, RoleVals(roleEntityKey, "Sales Rep"), nil)
	if err != nil {
		return err
	}
	// add role item - Support Engineer
	itemIDMap["support_engineer"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["support_engineer"], b.UserID, RoleVals(roleEntityKey, "Support Engineer"), nil)
	if err != nil {
		return err
	}
	// add role item - Director
	itemIDMap["director"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, rolesEntity.ID, itemIDMap["director"], b.UserID, RoleVals(roleEntityKey, "Director"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Roles Created")

	// add entity - payroll
	payrollEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityPayroll, "Payroll", entity.CategorySubData, entity.StateTeamLevel, false, false, false, PayrollFields())
	if err != nil {
		return err
	}

	// add entity - salary
	salaryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntitySalary, "Salary", entity.CategorySubData, entity.StateTeamLevel, false, false, false, SalaryFields())
	if err != nil {
		return err
	}

	// add entity - employees
	employeeEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmployee, "Employees", entity.CategoryData, entity.StateTeamLevel, false, true, false, EmployeeFields(b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id"), b.OwnerEntity.ID, ownerKey, rolesEntity.ID, roleEntityKey))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Employee Entity Created")
	// add entity - asset-catagory
	assetCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetCatagory, "Assets Category", entity.CategorySubData, entity.StateTeamLevel, false, false, false, forms.AssetCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Catagory Entity Created")
	// add entity - assets
	assetCatagoryTitleField := entity.TitleField(assetCatagoryEntity.FieldsIgnoreError())
	assetEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssets, "Assets", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, forms.AssetFields(assetCatagoryEntity.ID, assetCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Assets Entity Created")
	assetsFieldsNamedKeysMap := assetEntity.NamedKeys()
	// add asset item - Macbook Pro
	itemIDMap["macbook_pro"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, assetEntity.ID, itemIDMap["macbook_pro"], b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "Macbook Pro", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add asset item - iphone_14
	itemIDMap["iphone_14"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, assetEntity.ID, itemIDMap["iphone_14"], b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "iPhone 14", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add asset item - Macbook Air
	itemIDMap["macbook_air"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, assetEntity.ID, itemIDMap["macbook_air"], b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "Macbook Air", []interface{}{}), nil)
	if err != nil {
		return err
	}

	// add asset item - iMac 24'
	itemIDMap["imac"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, assetEntity.ID, itemIDMap["imac"], b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "iMac 24 Inch", []interface{}{}), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Assets Created")

	// add entity - asset status
	assetStatusEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityStatus, "Asset Status", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, AssetStatusFields())
	if err != nil {
		return err
	}
	statusFieldsNamedKeysMap := assetStatusEntity.NamedKeys()
	// add status item - Request-received
	itemIDMap["status_received"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, itemIDMap["status_received"], b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Request Received", "#fb667e"), nil)
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
	assetRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetRequest, "Asset Request", entity.CategorySubData, entity.StateTeamLevel, false, false, false, AssetRequestFields(assetEntity.ID, assetTitleField.Key, assetStatusEntity.ID, assetStatusTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Request Entity Created")

	// add entity - service-catagory
	serviceCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceCatagory, "Service Category", entity.CategorySubData, entity.StateTeamLevel, false, false, false, forms.ServiceCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Catagory Entity Created")
	// add entity - services
	serviceCatagoryTitleField := entity.TitleField(serviceCatagoryEntity.FieldsIgnoreError())
	serviceEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServices, "Services", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, forms.ServiceFields(serviceCatagoryEntity.ID, serviceCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Services Entity Created")
	servicesFieldsNamedKeysMap := serviceEntity.NamedKeys()
	// add service item - Git Access
	itemIDMap["git"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["git"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Git access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Mail Access
	itemIDMap["mail"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["mail"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Mail access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Slack Access
	itemIDMap["slack"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["slack"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Slack access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Infra Access
	itemIDMap["infra"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["infra"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Infra access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Marketing Tools Access
	itemIDMap["marketing"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["marketing"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Marketing tools access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	// add service item - Earnings Access
	itemIDMap["earnings"] = uuid.New().String()
	_, err = b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["earnings"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Earning sheet access", []interface{}{}), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Services Created")
	// add entity - service request
	serviceTitleField := entity.TitleField(serviceEntity.FieldsIgnoreError())
	serviceRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceRequest, "Service Request", entity.CategorySubData, entity.StateTeamLevel, false, false, false, ServiceRequestFields(serviceEntity.ID, serviceTitleField.Key, assetStatusEntity.ID, assetStatusTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Request Entity Created")

	err = AddAssociations(ctx, b, payrollEntity, salaryEntity, employeeEntity, assetRequestEntity, serviceRequestEntity)
	if err != nil {
		return err
	}
	err = AddSamples(ctx, b, itemIDMap)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	return nil
}

func AddAssociations(ctx context.Context, b *base.Base, payrollEid, salaryEid, employeeEid, assetRequestEid, serviceRequestEid entity.Entity) error {

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	//employee payroll association
	_, err = b.AssociationAdd(ctx, employeeEid.ID, payrollEid.ID)
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

	//payroll salary association
	_, err = b.AssociationAdd(ctx, payrollEid.ID, salaryEid.ID)
	if err != nil {
		return err
	}

	//employee stream association
	_, err = b.AssociationAdd(ctx, employeeEid.ID, streamEntity.ID)
	if err != nil {
		return err
	}

	//payroll stream association
	_, err = b.AssociationAdd(ctx, payrollEid.ID, streamEntity.ID)
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

func AddSamples(ctx context.Context, b *base.Base, itemIDMap map[string]string) error {

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	employeeEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmployee)
	if err != nil {
		return err
	}

	assetRequestEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAssetRequest)
	if err != nil {
		return err
	}

	serviceRequestEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityServiceRequest)
	if err != nil {
		return err
	}

	assetTemplateMacbookPro, err := b.TemplateAdd(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Asset request: Macbook Pro", itemIDMap["macbook_pro"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateMacbookAir, err := b.TemplateAdd(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Asset request: Macbook Air", itemIDMap["macbook_air"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateiPhone14, err := b.TemplateAdd(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Asset request: iPhone 14", itemIDMap["iphone_14"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateiMac, err := b.TemplateAdd(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Asset request: iMac", itemIDMap["imac"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateGit, err := b.TemplateAdd(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Service request: Git Access", itemIDMap["git"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateMail, err := b.TemplateAdd(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Service request: Mail Access", itemIDMap["mail"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateInfra, err := b.TemplateAdd(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Service request: Infra Access", itemIDMap["infra"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateMarketing, err := b.TemplateAdd(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Service request: Merketing Tools Access", itemIDMap["marketing"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateEarnings, err := b.TemplateAdd(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Service request: Earnings Sheet Access", itemIDMap["earnings"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate1, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare financial report for ", taskEntity, employeeEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate2, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Remove all the given access for ", taskEntity, employeeEntity), nil)
	if err != nil {
		return err
	}

	namedKeysMap := employeeEntity.NamedKeys()

	cp := &base.CoreWorkflow{
		Name:    "Employee Lifecycle",
		ActorID: employeeEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Onboarding",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Onboarding Employees",
				Nodes: []*base.CoreNode{
					{
						Name:       "Assign Macbook Pro",
						ActorID:    assetRequestEntity.ID,
						ActorName:  "Assest Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["developer"]),
						TemplateID: assetTemplateMacbookPro.ID,
						Tokens:     map[string]interface{}{itemIDMap["developer"]: "Developer"},
					},
					{
						Name:       "Assign Macbook Air",
						ActorID:    assetRequestEntity.ID,
						ActorName:  "Assest Request",
						Exp:        fmt.Sprintf("{{%s.%s}} !in {%s,%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["developer"], itemIDMap["designer"]),
						Tokens:     map[string]interface{}{itemIDMap["developer"]: "Developer", itemIDMap["designer"]: "Desingner"},
						TemplateID: assetTemplateMacbookAir.ID,
					},
					{
						Name:       "Assign iMac 24 inch",
						ActorID:    assetRequestEntity.ID,
						ActorName:  "Assest Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["designer"]),
						TemplateID: assetTemplateiMac.ID,
						Tokens:     map[string]interface{}{itemIDMap["designer"]: "Desingner"},
					},
					{
						Name:       "Assign iPhone 14",
						ActorID:    assetRequestEntity.ID,
						ActorName:  "Assest Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s,%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["support_engineer"], itemIDMap["sales_rep"]),
						TemplateID: assetTemplateiPhone14.ID,
						Tokens:     map[string]interface{}{itemIDMap["support_engineer"]: "Support Engineer", itemIDMap["sales_rep"]: "Sales Rep"},
					},
					{
						Name:       "Give Git access",
						ActorID:    serviceRequestEntity.ID,
						ActorName:  "Service Request",
						TemplateID: serviceTemplateGit.ID,
					},
					{
						Name:       "Give Mail access",
						ActorID:    serviceRequestEntity.ID,
						ActorName:  "Service Request",
						TemplateID: serviceTemplateMail.ID,
					},
					{
						Name:       "Give Infra access",
						ActorID:    serviceRequestEntity.ID,
						ActorName:  "Service Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["developer"]),
						TemplateID: serviceTemplateInfra.ID,
						Tokens:     map[string]interface{}{itemIDMap["developer"]: "Developer"},
					},
					{
						Name:       "Give Marketing tools access",
						ActorID:    serviceRequestEntity.ID,
						ActorName:  "Service Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s,%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["sales_rep"], itemIDMap["manager"]),
						TemplateID: serviceTemplateMarketing.ID,
						Tokens:     map[string]interface{}{itemIDMap["sales_rep"]: "Sales Rep", itemIDMap["manager"]: "Manager"},
					},
					{
						Name:       "Give Earnings report access",
						ActorID:    serviceRequestEntity.ID,
						ActorName:  "Service Request",
						Exp:        fmt.Sprintf("{{%s.%s}} in {%s}", employeeEntity.ID, namedKeysMap["role"], itemIDMap["director"]),
						TemplateID: serviceTemplateEarnings.ID,
						Tokens:     map[string]interface{}{itemIDMap["director"]: "Director"},
					},
					{
						Name:       "Prepare payroll",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate1.ID,
					},
				},
			},
			{
				Name:      "Exit",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Exit Employees",
				Nodes: []*base.CoreNode{
					{
						Name:       "Remove all access",
						ActorID:    taskEntity.ID,
						ActorName:  "Remove access",
						TemplateID: taskTemplate2.ID,
					},
				},
			},
		},
	}

	err = b.AddPipelines(ctx, cp)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:SAMPLES Pipeline And Its Nodes Created")

	inviteTemplate, err := b.TemplateAdd(ctx, b.InviteEntity.ID, uuid.New().String(), b.UserID, inviteTemplates("Hi, Welcome to the account", b.InviteEntity, employeeEntity), nil)
	if err != nil {
		return err
	}

	fmt.Println("\tEM:SAMPLES inviteTemplate added")

	cf := &base.CoreWorkflow{
		Name:    "When a new employee added",
		ActorID: employeeEntity.ID,
		Nodes: []*base.CoreNode{
			{
				Name:       "Invite Employee as users to the portal",
				ActorID:    b.InviteEntity.ID,
				ActorName:  "Employees",
				TemplateID: inviteTemplate.ID,
				Type:       node.Invite,
			},
		},
	}

	err = b.AddWorkflows(ctx, cf)
	if err != nil {
		return err
	}

	fmt.Println("\tEM:SAMPLES Workflows And Its Nodes Created")

	return nil
}
