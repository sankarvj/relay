package csm

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/dashboard"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	err := b.LoadFixedEntities(ctx)
	if err != nil {
		fmt.Println("\tCSP:BOOT LoadFixedEntities failed")
		return err
	}

	err = crm.CreateContactCompanyTaskEntity(ctx, b)
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
	fmt.Println("\tCRM:BOOT Approvals Entity Started")
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityApprovals, "Approvals", entity.CategoryApprovals, entity.StateTeamLevel, false, false, false, forms.ApprovalsFields(b.ApprovalStatusEntity.ID, b.ApprovalStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Approvals Entity Created")

	// add entity - activities
	b.ActivityEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityActivities, "Activities", entity.CategoryEvent, entity.StateAccountLevel, false, true, false, ActivitiesFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.ContactEntity.Key("email"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Activities Entity Created")

	// add entity - plan
	b.SubscriptionEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntitySubscriptions, "Subscriptions", entity.CategoryEvent, entity.StateAccountLevel, false, true, false, PlanFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
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
	err = b.AddSegments(ctx, b.ActivityEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, b.SubscriptionEntity.ID)
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
	//mark task as public
	entity.MarkAsPublic(ctx, b.AccountID, taskEntity.ID, true, b.DB, b.SecDB)

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
	fmt.Println("\tCSP:SAMPLES Events Created")

	err = addContacts(ctx, b, contactEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Contacts Items Created")

	err = addCompanies(ctx, b, companyEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Companies Item Created")

	err = addStreams(ctx, b, streamEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Streams Item Created")

	err = addProjects(ctx, b, projectEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Projects Item Created")

	err = addActivities(ctx, b, activityEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Activities Item Created")

	err = addSubscriptions(ctx, b, planEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Subscriptions Item Created")

	err = addDashboards(ctx, b, activityEntity, planEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Charts Created")

	return nil
}

func addEvents(ctx context.Context, b *base.Base) error {
	var err error
	b.DAUEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDailyActiveUsers, "Daily Active Users", entity.CategoryTimeSeries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcLatest, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityPageVisits, "Page Visits", entity.CategoryTimeSeries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcSum, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMilestones, "Milestones or Goals", entity.CategoryTimeSeries, entity.StateAccountLevel, false, false, false, events(entity.MetaCalcSum, entity.MetaRollUpAlways))
	if err != nil {
		return err
	}

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, proEid, emailEid, streamEID, taskEID, approvalsEID entity.Entity) error {

	var err error
	//contact company association
	_, err = b.AssociationAdd(ctx, b.ContactEntity.ID, b.CompanyEntity.ID)
	if err != nil {
		log.Println("ignoring error here. because the contraint might fail if it is alreay added in CRP")
		//return err
	}

	//task approvals association
	_, err = b.AssociationAdd(ctx, taskEID.ID, approvalsEID.ID)
	if err != nil {
		log.Println("ignoring error here. because the contraint might fail if it is alreay added in CRP")
		return err
	}

	//project email association
	// _, err = b.AssociationAdd(ctx, proEid.ID, emailEid.ID)
	// if err != nil {
	// 	return err
	// }

	//project task association
	_, err = b.AssociationAdd(ctx, proEid.ID, taskEID.ID)
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
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskVals(taskEntity, "Send demo link to the customer", "Send an email to the customer with the demo links often used to get started", b.ContactItemMatt.ID, b.StatusItemOpened.ID), map[string][]string{contactEntity.ID: {b.ContactItemMatt.ID}})
	if err != nil {
		return err
	}
	// add task item for contact - natasha (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskVals(taskEntity, "Schedule an on-site meeting with customer", "Ask the customer about the preferred time and schedule a meeting", b.ContactItemNatasha.ID, b.StatusItemOpened.ID), map[string][]string{contactEntity.ID: {b.ContactItemNatasha.ID}})
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:SAMPLES Tasks Items Created For Matt & Natasha")

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

func addStreams(ctx context.Context, b *base.Base, streamEntity entity.Entity) error {
	_, err := b.ItemAdd(ctx, streamEntity.ID, uuid.New().String(), b.UserID, forms.StreamVals(streamEntity, "General", "conversations"), nil)
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

func addDashboards(ctx context.Context, b *base.Base, activityEntity, planEntity, approvalsEntity entity.Entity) error {

	homeDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, entity.NoEntityID, "Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addHomeCharts(ctx, b, homeDashID, activityEntity, planEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard Overview Created")
	projDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, b.ProjectEntity.ID, "Project Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addProjectCharts(ctx, b, projDashID, activityEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard Project Overview Created")
	myDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, b.NotificationEntity.ID, "My Dashboard").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addMyCharts(ctx, b, myDashID, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCSP:Dashboard My Dashboard Created")

	return nil
}

func addHomeCharts(ctx context.Context, b *base.Base, dashboardID string, activityEntity, planEntity entity.Entity) error {
	//charts for home dashboard
	var err error
	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.DAUEntity.ID, "daily_active_users", "Daily active users", "", chart.TypeLine).SetAsTimeseries().Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ContactEntity.ID, "contacts_stage", "Contacts stage", "lifecycle_stage", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.CompanyEntity.ID, "accounts_health", "Accounts health", "health", chart.TypeBar).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	// err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.DAUEntity.ID, "dau", "DAU", "", chart.TypeGrid).SetAsTimeseries().SetDurationLast24hrs().SetCalcSum().SetIcon("dau.svg").Add(ctx, b.DB)
	// if err != nil {
	// 	return err
	// }
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ContactEntity.ID, "chrun_rate", "Chrun rate", "lifecycle_stage", chart.TypeGrid).AddDateField("lost_customer_on").SetDurationLast24hrs().SetCalcRate().SetGrpLogicID().SetIcon("chrun.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ContactEntity.ID, "new_customer", "New Customer", "", chart.TypeGrid).AddDateField("became_a_customer_date").SetDurationLast24hrs().SetCalcSum().SetIcon("new_customer.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	exp1 := fmt.Sprintf("{{%s.%s}} bf {%s}", b.ProjectEntity.ID, b.ProjectEntity.Key("end_time"), "now")
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "delayed", "Delayed", "end_time", chart.TypeGrid).AddExp(exp1).AddSource(b.CompanyEntity.ID).SetDurationLast24hrs().SetCalcSum().SetGrpLogicParent().SetIcon("delayed.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, activityEntity.ID, "goals", "Activities", "name", chart.TypeRod).SetDurationAllTime().SetGrpLogicField().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, planEntity.ID, "cancellations", "Cancellations", "reason", chart.TypeRod).SetDurationAllTime().SetGrpLogicID().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	return nil
}

func addProjectCharts(ctx context.Context, b *base.Base, dashboardID string, activityEntity entity.Entity) error {

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
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "stage", "Stage", b.ProjectEntity.Key("pipeline_stage"), chart.TypeGrid).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "status", "Status", b.ProjectEntity.Key("status"), chart.TypeGrid).SetAsCustom().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	advActivityMap := map[string]string{
		"associated_companies": activityEntity.Key("associated_companies"),
		"associated_contacts":  activityEntity.Key("associated_contacts"),
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, activityEntity.ID, "goals", "Goals", "name", chart.TypeRod).
		AddAdvancedMap(advActivityMap).
		SetDurationAllTime().
		SetGrpLogicField().
		Add(ctx, b.DB)
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

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "project_phase", "Project Phase", "pipeline_stage", chart.TypePie).
		SetDurationAllTime().
		SetGrpLogicID().
		Add(ctx, b.DB)
	if err != nil {
		return err
	}

	return nil
}

func addMyCharts(ctx context.Context, b *base.Base, dashboardID string, approvalsEntity entity.Entity) error {
	//charts for notifications me
	paOnMeExp := fmt.Sprintf("{{%s.%s}} in {%s,%s} && {{%s.%s}} in {{me}}", approvalsEntity.ID, approvalsEntity.Key("status"), b.ApprovalStatusWaiting.ID, b.ApprovalStatusChangeRequested.ID, approvalsEntity.ID, approvalsEntity.Key("assignees"))
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, approvalsEntity.ID, "my_pending_approvals", "My Pending Approvals", "", chart.TypeCard).AddExp(paOnMeExp).SetDurationAllTime().SetGrpLogicID().SetIcon("pending-approvals.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	overdueOnMeEXP := fmt.Sprintf("{{%s.%s}} !eq {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemClosed.ID, b.TaskEntity.ID, b.TaskEntity.Key("due_by"), "now", b.TaskEntity.ID, b.TaskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "my_overdue_tasks", "My Overdue Tasks", "", chart.TypeCard).AddExp(overdueOnMeEXP).SetDurationAllTime().SetIcon("overdue-tasks.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	openOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", b.TaskEntity.ID, b.TaskEntity.Key("status"), b.StatusItemOpened.ID, b.TaskEntity.ID, b.TaskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.TaskEntity.ID, "my_open_tasks", "My Open Tasks", "", chart.TypeCard).AddExp(openOnMeEXP).SetDurationAllTime().SetIcon("open-tasks.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	overdueProjOnMeEXP := fmt.Sprintf("{{%s.%s}} !in {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", b.ProjectEntity.ID, b.ProjectEntity.Key("status"), b.StatusItemClosed.ID, b.ProjectEntity.ID, b.ProjectEntity.Key("end_time"), "now", b.ProjectEntity.ID, b.ProjectEntity.Key("owner"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "my_overdue_projects", "My Overdue Projects", "", chart.TypeCard).AddExp(overdueProjOnMeEXP).SetDurationAllTime().SetIcon("overdue-projects.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	openProjOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", b.ProjectEntity.ID, b.ProjectEntity.Key("status"), b.StatusItemOpened.ID, b.ProjectEntity.ID, b.ProjectEntity.Key("owner"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, b.ProjectEntity.ID, "my_open_projects", "My Open Projects", "", chart.TypeCard).AddExp(openProjOnMeEXP).SetDurationAllTime().SetIcon("open-projects.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}

	return nil
}
