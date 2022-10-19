package csm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)

	err := crm.CreateContactCompanyTaskEntity(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT ConComTask Entity Created")

	// add entity - project
	b.ProjectEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityProjects, "Projects", entity.CategoryData, entity.StateTeamLevel, false, true, false, ProjectFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"), b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Projects Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, false, false, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.ProjectEntity.ID, b.ContactEntity.Key("email"), b.ContactEntity.Key("first_name"), b.CompanyEntity.Key("name"), b.ProjectEntity.Key("project_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - approvals
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityApprovals, "Approvals", entity.CategoryApprovals, entity.StateTeamLevel, false, false, false, forms.ApprovalsFields(b.ApprovalStatusEntity.ID, b.ApprovalStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Approvals Entity Created")

	// add entity - activities
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityActivities, "Activities", entity.CategoryEvent, entity.StateAccountLevel, false, false, false, ActivitiesFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Activities Entity Created")

	// add entity - plan
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntitySubscriptions, "Subscriptions", entity.CategoryEvent, entity.StateAccountLevel, false, false, false, PlanFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Subscriptions Entity Created")

	return nil

}

func AddWorkflows(ctx context.Context, b *base.Base) error {
	err := addAutomation(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Automations Created")

	err = b.AddSegments(ctx, b.ProjectEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, b.ContactEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, b.CompanyEntity.ID)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Sample Segments Created For Contacts/Companies/Deals")
	return nil
}

func AddSamples(ctx context.Context, b *base.Base) error {
	contactEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityContacts)
	if err != nil {
		return err
	}
	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityCompanies)
	if err != nil {
		return err
	}
	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityProjects)
	if err != nil {
		return err
	}

	activityEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityActivities)
	if err != nil {
		return err
	}
	planEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntitySubscriptions)
	if err != nil {
		return err
	}
	emailsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmails)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	approvalsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityApprovals)
	if err != nil {
		return err
	}

	err = addAssociations(ctx, b, projectEntity, emailsEntity, streamEntity, taskEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	err = addEvents(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Events Created")

	err = addContacts(ctx, b, contactEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Contacts Items Created")

	err = addCompanies(ctx, b, companyEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Companies Item Created")

	err = addProjects(ctx, b, projectEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Projects Item Created")

	err = addActivities(ctx, b, activityEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Activities Item Created")

	err = addSubscriptions(ctx, b, planEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Subscriptions Item Created")

	err = addCharts(ctx, b, activityEntity, planEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Charts Created")

	return nil
}

func addEvents(ctx context.Context, b *base.Base) error {
	var err error
	b.DAUEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDailyActiveUsers, "Daily Active Users", entity.CategoryTimeseries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcLatest, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityPageVisits, "Page Visits", entity.CategoryTimeseries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcSum, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMilestones, "Milestones or Goals", entity.CategoryTimeseries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcSum, entity.MetaRollUpAlways))
	if err != nil {
		return err
	}

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, proEid, emailEid, streamEID, taskEID, approvalsEID entity.Entity) error {

	//project email association
	_, err := b.AssociationAdd(ctx, proEid.ID, emailEid.ID)
	if err != nil {
		return err
	}

	//project task association
	_, err = b.AssociationAdd(ctx, proEid.ID, taskEID.ID)
	if err != nil {
		return err
	}

	//task approvals association
	_, err = b.AssociationAdd(ctx, taskEID.ID, approvalsEID.ID)
	if err != nil {
		return err
	}

	//ASSOCIATE STREAMS
	//project stream association
	_, err = b.AssociationAdd(ctx, proEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	return nil
}

func addContacts(ctx context.Context, b *base.Base, contactEntity, taskEntity entity.Entity) error {
	var err error
	// add contact item
	b.ContactItemMatt, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Matt", "Murdock", "matt@starkindst.com", b.LeadStatusItemNew.ID), nil)
	if err != nil {
		return err
	}
	// add contact item
	b.ContactItemNatasha, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Natasha", "Romanova", "natasha@randcorp.com", b.LeadStatusItemConnected.ID), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bruce", "Banner", "bruce@alumina.com", b.LeadStatusItemAttempted.ID), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bucky", "Barnes", "bucky@dailybugle.com", b.LeadStatusItemBadTiming.ID), nil)
	if err != nil {
		return err
	}

	// add task item for contact - matt (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskVals(taskEntity, "Send demo link to the customer", b.ContactItemMatt.ID), map[string][]string{contactEntity.ID: {b.ContactItemMatt.ID}})
	if err != nil {
		return err
	}
	// add task item for contact - natasha (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskVals(taskEntity, "Schedule an on-site meeting with customer", b.ContactItemNatasha.ID), map[string][]string{contactEntity.ID: {b.ContactItemNatasha.ID}})
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Tasks Items Created For Matt & Natasha")

	return nil
}

func addCompanies(ctx context.Context, b *base.Base, companyEntity entity.Entity) error {
	var err error
	b.CompanyItemStarkInd, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Stark Industries", "starkindst.com"), nil)
	if err != nil {
		return err
	}

	b.CompanyItemRandInd, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Rand corporation", "randcorp.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Alumina", "alumina.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Daily bugle", "dailybugle.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Salesforce Inc", "salesforce.com"), nil)
	if err != nil {
		return err
	}

	return nil
}

func addProjects(ctx context.Context, b *base.Base, projectEntity, contactEntity entity.Entity) error {
	_, err := b.ItemAdd(ctx, projectEntity.ID, uuid.New().String(), b.UserID, ProjVals(projectEntity, "Base Project", b.ContactItemMatt.ID, b.ContactItemNatasha.ID, b.SalesPipelineFlowID), nil)
	if err != nil {
		return err
	}
	return nil
}

func addActivities(ctx context.Context, b *base.Base, activityEntity, contactEntity entity.Entity) error {

	_, err := b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Flow Completed", "All the flows completed", b.ContactItemMatt.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Checked Subscription Page", "Customer viewed subscription page", b.ContactItemMatt.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Encountered Error", "Customer encountered an error", b.ContactItemMatt.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Invoice Created", "Invoice feature is clicked and visited", b.ContactItemNatasha.ID, b.CompanyItemRandInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Invoice Created", "Invoice feature is clicked and visited", b.ContactItemMatt.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Members Invited", "Few members are invited", b.ContactItemNatasha.ID, b.CompanyItemRandInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Data Populated", "Data populated for the first time", b.ContactItemNatasha.ID, b.CompanyItemRandInd.ID), nil)
	if err != nil {
		return err
	}
	_, err = b.ItemAdd(ctx, activityEntity.ID, uuid.New().String(), b.UserID, ActivitiesVals(activityEntity, "Data Populated", "Data populated for the first time", b.ContactItemMatt.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	return nil
}

func addSubscriptions(ctx context.Context, b *base.Base, subscriptionEntity, contactEntity entity.Entity) error {
	_, err := b.ItemAdd(ctx, subscriptionEntity.ID, uuid.New().String(), b.UserID, PlanVals(subscriptionEntity, "Lost a customer", "Lost a customer due to technical issues", b.ContactItemMatt.ID, b.ContactItemNatasha.ID, b.CompanyItemStarkInd.ID), nil)
	if err != nil {
		return err
	}
	return nil
}

func addCharts(ctx context.Context, b *base.Base, activityEntity, planEntity entity.Entity) error {

	err := chart.BuildNewChart(b.AccountID, b.UserID, b.DAUEntity.ID, "Daily active users", "", chart.TypeLine).SetAsTimeseries().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ContactEntity.ID, "Contacts by stage", "lifecycle_stage", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.CompanyEntity.ID, "Accounts by health bar", "health", chart.TypeBar).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.UserID, b.DAUEntity.ID, "DAU", "", chart.TypeGrid).SetAsTimeseries().SetDurationLast24hrs().SetCalcSum().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ContactEntity.ID, "Chrun rate", "lifecycle_stage", chart.TypeGrid).AddDateField("lost_customer_on").SetDurationLast24hrs().SetCalcRate().SetGrpLogicID().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ContactEntity.ID, "New Customer", "", chart.TypeGrid).AddDateField("became_a_customer_date").SetDurationLast24hrs().SetCalcSum().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	exp1 := fmt.Sprintf("{{%s.%s}} bf {%s}", b.ProjectEntity.ID, b.ProjectEntity.Key("end_time"), "now")
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ProjectEntity.ID, "Delayed Accounts", "end_time", chart.TypeGrid).AddExp(exp1).AddSource(b.CompanyEntity.ID).SetDurationLast24hrs().SetCalcSum().SetGrpLogicParent().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.UserID, activityEntity.ID, "Activities", "activity_name", chart.TypeRod).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, planEntity.ID, "Cancellations", "reason", chart.TypeRod).SetDurationAllTime().SetGrpLogicID().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	//Charts for project
	overdueEXP := fmt.Sprintf("{{%s.%s}} !in {%s} && {{%s.%s}} bf {%s}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemClosed.ID, b.TaskEntity.ID, b.TaskEntity.Key("due_by"), "now")
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.TaskEntity.ID, "Overdue", "", chart.TypeGrid).AddExp(overdueEXP).SetBaseEntityID(b.ProjectEntity.ID).SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	openEXP := fmt.Sprintf("{{%s.%s}} in {%s}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemOpened.ID)
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.TaskEntity.ID, "Open", "", chart.TypeGrid).AddExp(openEXP).SetBaseEntityID(b.ProjectEntity.ID).SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ProjectEntity.ID, "Stage", b.ProjectEntity.Key("pipeline_stage"), chart.TypeGrid).SetBaseEntityID(b.ProjectEntity.ID).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.ProjectEntity.ID, "Status", b.ProjectEntity.Key("status"), chart.TypeGrid).SetBaseEntityID(b.ProjectEntity.ID).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.UserID, activityEntity.ID, "Goals", "activity_name", chart.TypeRod).SetBaseEntityID(b.ProjectEntity.ID).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.TaskEntity.ID, "Tasks", "status", chart.TypePie).SetBaseEntityID(b.ProjectEntity.ID).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.UserID, b.TaskEntity.ID, "Overdue Tasks", "", chart.TypeList).AddExp(overdueEXP).SetBaseEntityID(b.ProjectEntity.ID).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	return nil
}
