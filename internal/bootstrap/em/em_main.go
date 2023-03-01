package em

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/dashboard"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)
	b.OwnerEntity.Fields()
	ownerKey := b.OwnerEntity.NameKeyMapWrapper()["email"]
	fmt.Println("\tEM:BOOT Owner Entity Retrived")

	itemIDMap := make(map[string]string, 0)

	// add entity - roles
	rolesEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityRoles, "Roles", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, RoleFields())
	if err != nil {
		return err
	}

	// add entity - payroll
	payrollEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityPayroll, "Payroll", entity.CategoryData, entity.StateTeamLevel, false, false, false, PayrollFields())
	if err != nil {
		return err
	}

	// add entity - salary
	salaryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntitySalary, "Salary", entity.CategoryData, entity.StateTeamLevel, false, false, false, SalaryFields())
	if err != nil {
		return err
	}

	// add entity - employees
	employeeEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmployee, "Employees", entity.CategoryData, entity.StateTeamLevel, false, true, false, EmployeeFields(b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id"), b.OwnerEntity.ID, ownerKey, rolesEntity.ID, rolesEntity.Key("role")))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Employee Entity Created")

	// add entity - tasks
	taskEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityETask, "Tasks", entity.CategoryTask, entity.StateTeamLevel, false, true, true, TaskEFields(employeeEntity.ID, employeeEntity.Key("first_name"), b.NodeEntity.ID, b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, ownerKey))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Employee Task Entity Created")

	// add entity - approvals
	fmt.Println("\tCRM:BOOT Approvals Entity Started")
	approvalEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityApprovals, "Approvals", entity.CategoryApprovals, entity.StateTeamLevel, false, false, false, forms.ApprovalsFields(b.ApprovalStatusEntity.ID, b.ApprovalStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Approvals Entity Created")

	// add entity - asset-catagory
	assetCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetCatagory, "Assets Category", entity.CategoryData, entity.StateTeamLevel, false, false, false, forms.AssetCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Catagory Entity Created")
	// add entity - assets
	assetCatagoryTitleField := entity.TitleField(assetCatagoryEntity.EasyFields())
	assetEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssets, "Assets", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, forms.AssetFields(assetCatagoryEntity.ID, assetCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Assets Entity Created")

	// add entity - asset status
	assetStatusEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetStatus, "Asset Status", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, AssetStatusFields())
	if err != nil {
		return err
	}

	// add entity - asset request
	assetRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAssetRequest, "Asset Requests", entity.CategoryTask, entity.StateTeamLevel, false, true, false, AssetRequestFields(assetEntity.ID, assetEntity.Key("asset_name"), assetStatusEntity.ID, assetStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Request Entity Created")

	// add entity - service-catagory
	serviceCatagoryEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceCatagory, "Service Category", entity.CategoryData, entity.StateTeamLevel, false, false, false, forms.ServiceCategory())
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Catagory Entity Created")
	// add entity - services
	serviceCatagoryTitleField := entity.TitleField(serviceCatagoryEntity.EasyFields())
	serviceEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServices, "Services", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, forms.ServiceFields(serviceCatagoryEntity.ID, serviceCatagoryTitleField.Key))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Services Entity Created")

	// add entity - service request
	serviceTitleField := entity.TitleField(serviceEntity.EasyFields())
	serviceRequestEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityServiceRequest, "Service Requests", entity.CategoryTask, entity.StateTeamLevel, false, true, false, ServiceRequestFields(serviceEntity.ID, serviceTitleField.Key, assetStatusEntity.ID, assetStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Service Request Entity Created")

	err = AddAssociations(ctx, b, payrollEntity, salaryEntity, employeeEntity, assetRequestEntity, serviceRequestEntity, taskEntity, approvalEntity)
	if err != nil {
		return err
	}
	err = AddSamples(ctx, b, itemIDMap)
	if err != nil {
		return err
	}
	err = AddWorkflows(ctx, b, employeeEntity, taskEntity, assetRequestEntity, serviceRequestEntity, itemIDMap)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	return nil
}

func AddAssociations(ctx context.Context, b *base.Base, payrollEid, salaryEid, employeeEid, assetRequestEid, serviceRequestEid, taskEntity, approvalsEntity entity.Entity) error {

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

	//task approvals association
	_, err = b.AssociationAdd(ctx, taskEntity.ID, approvalsEntity.ID)
	if err != nil {
		return err
	}

	//asset-request approvals association
	_, err = b.AssociationAdd(ctx, assetRequestEid.ID, approvalsEntity.ID)
	if err != nil {
		return err
	}

	//service-request approvals association
	_, err = b.AssociationAdd(ctx, serviceRequestEid.ID, approvalsEntity.ID)
	if err != nil {
		return err
	}

	return nil
}

func AddSamples(ctx context.Context, b *base.Base, itemIDMap map[string]string) error {

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityETask)
	if err != nil {
		return err
	}
	b.TaskEntity = taskEntity

	employeeEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmployee)
	if err != nil {
		return err
	}

	rolesEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityRoles)
	if err != nil {
		return err
	}
	assetStatusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAssetStatus)
	if err != nil {
		return err
	}

	assetsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAssets)
	if err != nil {
		return err
	}
	servicesEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityServices)
	if err != nil {
		return err
	}

	serviceRequestEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityServiceRequest)
	if err != nil {
		return err
	}
	assetRequestEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAssetRequest)
	if err != nil {
		return err
	}

	approvalEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityApprovals)
	if err != nil {
		return err
	}

	err = addRoles(ctx, b, itemIDMap, rolesEntity)
	if err != nil {
		return err
	}

	err = addAssets(ctx, b, employeeEntity, assetsEntity, itemIDMap)
	if err != nil {
		return err
	}

	err = addStatuses(ctx, b, employeeEntity, assetStatusEntity, itemIDMap)
	if err != nil {
		return err
	}

	err = addServices(ctx, b, employeeEntity, servicesEntity, itemIDMap)
	if err != nil {
		return err
	}

	err = addEmployees(ctx, b, employeeEntity, taskEntity, itemIDMap)
	if err != nil {
		return err
	}

	err = addDashboards(ctx, b, employeeEntity, serviceRequestEntity, assetRequestEntity, approvalEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Charts Created")

	return nil
}

func addRoles(ctx context.Context, b *base.Base, itemIDMap map[string]string, rolesEntity entity.Entity) error {
	var err error
	roleEntityKey := rolesEntity.NameKeyMapWrapper()["role"]
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
	return nil
}

func addAssets(ctx context.Context, b *base.Base, empEntity, assetEntity entity.Entity, itemIDMap map[string]string) error {
	assetsFieldsNamedKeysMap := assetEntity.NameKeyMapWrapper()
	// add asset item - Macbook Pro
	itemIDMap["macbook_pro"] = uuid.New().String()
	_, err := b.ItemAdd(ctx, assetEntity.ID, itemIDMap["macbook_pro"], b.UserID, forms.AssetVals(assetsFieldsNamedKeysMap["asset_name"], assetsFieldsNamedKeysMap["catagory"], "Macbook Pro", []interface{}{}), nil)
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
	return nil
}

func addStatuses(ctx context.Context, b *base.Base, empEntity, assetStatusEntity entity.Entity, itemIDMap map[string]string) error {
	statusFieldsNamedKeysMap := assetStatusEntity.NameKeyMapWrapper()
	// add status item - Request-received
	itemIDMap["status_received"] = uuid.New().String()
	_, err := b.ItemAdd(ctx, assetStatusEntity.ID, itemIDMap["status_received"], b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Request Received", "#FFE9AE"), nil)
	if err != nil {
		return err
	}
	// add status item - In-progress
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "In Progress", "#FFEF82"), nil)
	if err != nil {
		return err
	}
	// add status item - Approved
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Approved", "#B4E197"), nil)
	if err != nil {
		return err
	}
	// add status item - Delivered
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Delivered", "#00ADB5"), nil)
	if err != nil {
		return err
	}
	// add status item - Returned
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Returned", "#FF8C8C"), nil)
	if err != nil {
		return err
	}
	// add status item - Trashed
	_, err = b.ItemAdd(ctx, assetStatusEntity.ID, uuid.New().String(), b.UserID, AssetStatusVals(statusFieldsNamedKeysMap["name"], statusFieldsNamedKeysMap["color"], "Trashed", "#E8E8E8"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tEM:BOOT Asset Status Entity Created")
	return nil
}

func addServices(ctx context.Context, b *base.Base, empEntity, serviceEntity entity.Entity, itemIDMap map[string]string) error {
	servicesFieldsNamedKeysMap := serviceEntity.NameKeyMapWrapper()
	// add service item - Git Access
	itemIDMap["git"] = uuid.New().String()
	_, err := b.ItemAdd(ctx, serviceEntity.ID, itemIDMap["git"], b.UserID, forms.AssetVals(servicesFieldsNamedKeysMap["service_name"], servicesFieldsNamedKeysMap["catagory"], "Git access", []interface{}{}), nil)
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
	return nil
}

func addEmployees(ctx context.Context, b *base.Base, empEntity, taskEntity entity.Entity, itemIDMap map[string]string) error {
	var err error
	// add employee item
	b.EmpItemMatt, err = b.ItemAdd(ctx, empEntity.ID, uuid.New().String(), b.UserID, EmpVals(empEntity, "Matt", "Murdock", "matt@starkindst.com", itemIDMap["intern"]), nil)
	if err != nil {
		return err
	}
	// add employee item
	b.EmpItemNatasha, err = b.ItemAdd(ctx, empEntity.ID, uuid.New().String(), b.UserID, EmpVals(empEntity, "Natasha", "Romanova", "natasha@randcorp.com", itemIDMap["developer"]), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, empEntity.ID, uuid.New().String(), b.UserID, EmpVals(empEntity, "Bruce", "Banner", "bruce@alumina.com", itemIDMap["designer"]), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, empEntity.ID, uuid.New().String(), b.UserID, EmpVals(empEntity, "Bucky", "Barnes", "bucky@dailybugle.com", itemIDMap["manager"]), nil)
	if err != nil {
		return err
	}

	fmt.Println("\tEMP:SAMPLES Employess Created")

	return nil
}

func AddWorkflows(ctx context.Context, b *base.Base, employeeEntity, taskEntity, assetRequestEntity, serviceRequestEntity entity.Entity, itemIDMap map[string]string) error {
	assetTemplateMacbookPro, err := b.TemplateAddWithOutMeta(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Macbook Pro", itemIDMap["macbook_pro"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateMacbookAir, err := b.TemplateAddWithOutMeta(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("Macbook Air", itemIDMap["macbook_air"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateiPhone14, err := b.TemplateAddWithOutMeta(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("iPhone 14", itemIDMap["iphone_14"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	assetTemplateiMac, err := b.TemplateAddWithOutMeta(ctx, assetRequestEntity.ID, uuid.New().String(), b.UserID, assetRequestTemplates("iMac", itemIDMap["imac"], itemIDMap["status_received"], assetRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateGit, err := b.TemplateAddWithOutMeta(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Git access", itemIDMap["git"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateMail, err := b.TemplateAddWithOutMeta(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Mail access", itemIDMap["mail"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateInfra, err := b.TemplateAddWithOutMeta(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Infra access", itemIDMap["infra"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateMarketing, err := b.TemplateAddWithOutMeta(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Marketing tools", itemIDMap["marketing"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	serviceTemplateEarnings, err := b.TemplateAddWithOutMeta(ctx, serviceRequestEntity.ID, uuid.New().String(), b.UserID, serviceRequestTemplates("Earnings sheet access", itemIDMap["earnings"], itemIDMap["status_received"], serviceRequestEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate1, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare financial report for ", taskEntity, employeeEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate2, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Remove all the given access for ", taskEntity, employeeEntity), nil)
	if err != nil {
		return err
	}

	namedKeysMap := employeeEntity.NameKeyMapWrapper()
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

	// inviteTemplate, err := b.TemplateAddWithOutMeta(ctx, b.InviteEntity.ID, uuid.New().String(), b.UserID, inviteTemplates("Hi, Welcome to the account", b.InviteEntity, employeeEntity), nil)
	// if err != nil {
	// 	return err
	// }

	fmt.Println("\tEM:SAMPLES inviteTemplate added")

	// cf := &base.CoreWorkflow{
	// 	Name:     "When a new employee added",
	// 	ActorID:  employeeEntity.ID,
	// 	FlowType: flow.FlowTypeEventCreate,
	// 	Nodes: []*base.CoreNode{
	// 		{
	// 			Name:       "Invite Employee as users to the portal",
	// 			ActorID:    b.InviteEntity.ID,
	// 			ActorName:  "Employees",
	// 			TemplateID: inviteTemplate.ID,
	// 			Type:       node.Invite,
	// 		},
	// 	},
	// }

	// err = b.AddWorkflows(ctx, cf)
	// if err != nil {
	// 	return err
	// }

	fmt.Println("\tEM:SAMPLES Workflows And Its Nodes Created")
	return nil
}

func addDashboards(ctx context.Context, b *base.Base, empEntity, srEntity, arEntity, approvalsEntity entity.Entity) error {

	homeDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, entity.NoEntityID, "Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addHomeCharts(ctx, b, homeDashID, empEntity, srEntity, arEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard Overview Created")
	projDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, empEntity.ID, "Employee Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addProjectCharts(ctx, b, projDashID, empEntity, srEntity, arEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard Project Overview Created")
	myDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, b.NotificationEntity.ID, "My Dashboard").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addMyCharts(ctx, b, myDashID, empEntity, srEntity, arEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard My Dashboard Created")

	return nil
}

func addHomeCharts(ctx context.Context, b *base.Base, dashboardID string, empEntity, srEntity, arEntity entity.Entity) error {
	//charts for home dashboard
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "employee_stage", "Employee stage", "lifecycle_stage", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "employee_roles", "Employee roles", "role", chart.TypeBar).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "exit", "Exit rate", "lifecycle_stage", chart.TypeGrid).AddDateField("exit_date").SetDurationLast24hrs().SetCalcRate().SetGrpLogicID().SetIcon("face-with-_x_-eyes-.png").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "join", "New Joinees", "", chart.TypeGrid).AddDateField("joining_date").SetDurationLast24hrs().SetCalcSum().SetIcon("spotted-sweater-girl-with-wand-torso.png").Add(ctx, b.DB)
	if err != nil {
		return err
	}

	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, activityEntity.ID, "goals", "Activities", "name", chart.TypeRod).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }
	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, planEntity.ID, "cancellations", "Cancellations", "reason", chart.TypeRod).SetDurationAllTime().SetGrpLogicID().Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func addProjectCharts(ctx context.Context, b *base.Base, dashboardID string, empEntity, srEntity, arEntity entity.Entity) error {

	//charts for projects
	overdueEXP := fmt.Sprintf("{{%s.%s}} !in {%s} && {{%s.%s}} bf {%s}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemClosed.ID, b.TaskEntity.ID, b.TaskEntity.Key("due_by"), "now")
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "overdue", "Overdue", "", chart.TypeGrid).AddExp(overdueEXP).SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	openEXP := fmt.Sprintf("{{%s.%s}} in {%s}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemOpened.ID)
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "open", "Open", "", chart.TypeGrid).AddExp(openEXP).SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "stage", "Stage", empEntity.Key("lifecyle_stage"), chart.TypeGrid).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, empEntity.ID, "status", "Status", empEntity.Key("role"), chart.TypeGrid).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "tasks", "Tasks", "status", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "overdue_tasks", "Overdue Tasks", "", chart.TypeList).AddExp(overdueEXP).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	return nil
}

func addMyCharts(ctx context.Context, b *base.Base, dashboardID string, empEntity, srEntity, arEntity, approvalsEntity entity.Entity) error {
	//charts for notifications me
	paOnMeExp := fmt.Sprintf("{{%s.%s}} in {%s,%s} && {{%s.%s}} in {{me}}", approvalsEntity.ID, approvalsEntity.Key("status"), b.ApprovalStatusWaiting.ID, b.ApprovalStatusChangeRequested.ID, approvalsEntity.ID, approvalsEntity.Key("assignees"))
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, approvalsEntity.ID, "my_pending_approvals", "My Pending Approvals", "", chart.TypeCard).AddExp(paOnMeExp).SetDurationAllTime().SetGrpLogicID().SetIcon("vertical-traffic-lights.png").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	overdueOnMeEXP := fmt.Sprintf("{{%s.%s}} !eq {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemClosed.ID, b.TaskEntity.ID, b.TaskEntity.Key("due_by"), "now", b.TaskEntity.ID, b.TaskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "my_overdue_tasks", "My Overdue Tasks", "", chart.TypeCard).AddExp(overdueOnMeEXP).SetDurationAllTime().SetIcon("round-wall-clock-yellow.png").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	openOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemOpened.ID, b.TaskEntity.ID, b.TaskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "my_open_tasks", "My Open Tasks", "", chart.TypeCard).AddExp(openOnMeEXP).SetDurationAllTime().SetIcon("round-wall-clock.png").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	// overdueProjOnMeEXP := fmt.Sprintf("{{%s.%s}} !in {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", b.ProjectEntity.ID, b.ProjectEntity.Key("status"), b.StatusItemClosed.ID, b.ProjectEntity.ID, b.ProjectEntity.Key("end_time"), "now", b.ProjectEntity.ID, b.ProjectEntity.Key("owner"))
	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "my_overdue_projects", "My Overdue Projects", "", chart.TypeCard).AddExp(overdueProjOnMeEXP).SetDurationAllTime().SetIcon("timetable-icon.png").Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }
	// openProjOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", b.ProjectEntity.ID, b.ProjectEntity.Key("status"), b.StatusItemOpened.ID, b.ProjectEntity.ID, b.ProjectEntity.Key("owner"))
	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "my_open_projects", "My Open Projects", "", chart.TypeCard).AddExp(openProjOnMeEXP).SetDurationAllTime().SetIcon("aim-board-with-stand.png").Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }

	return nil
}
