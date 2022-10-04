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
	b.ProjectEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityProjects, "Projects", entity.CategoryData, entity.StateTeamLevel, false, true, false, ProjectFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("email"), b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Projects Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, true, false, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.ProjectEntity.ID, b.ContactEntity.Key("email"), b.ContactEntity.Key("first_name"), b.CompanyEntity.Key("name"), b.ProjectEntity.Key("project_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - activities
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityActivities, "Activities", entity.CategoryEvent, entity.StateAccountLevel, false, true, false, ActivitiesFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Activities Entity Created")

	// add entity - plan
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntitySubscriptions, "Subscriptions", entity.CategoryEvent, entity.StateAccountLevel, false, true, false, PlanFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name")))
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

	err = addAssociations(ctx, b, projectEntity, emailsEntity, streamEntity, taskEntity)
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

	err = addCharts(ctx, b, activityEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Charts Created")

	return nil
}

func addEvents(ctx context.Context, b *base.Base) error {
	_, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDailyActiveUsers, "Daily Active Users", entity.CategoryTimeseries, entity.StateAccountLevel, false, true, false, events(entity.MetaCalcLatest, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityPageVisits, "Page Visits", entity.CategoryTimeseries, entity.StateAccountLevel, false, true, false, events(entity.MetaCalcSum, entity.MetaRollUpDaily))
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMilestones, "Milestones or Goals", entity.CategoryTimeseries, entity.StateAccountLevel, false, true, false, events(entity.MetaCalcSum, entity.MetaRollUpAlways))
	if err != nil {
		return err
	}

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, proEid, emailEid, streamEID, taskEID entity.Entity) error {

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

func addCharts(ctx context.Context, b *base.Base, dauEntity entity.Entity) error {
	NoEntityID := "00000000-0000-0000-0000-000000000000"
	err := b.ChartAdd(ctx, b.ContactEntity.ID, NoEntityID, "Contacts by stage", "lifecycle_stage", "pie", "normal", "last_week", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, b.CompanyEntity.ID, NoEntityID, "Accounts by health", "health", "pie", "normal", "last_week", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, b.CompanyEntity.ID, NoEntityID, "Accounts by health", "health", "bar", "normal", "last_week", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, dauEntity.ID, NoEntityID, "Daily active users", "", "line", "timeseries", "last_week", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, dauEntity.ID, NoEntityID, "Daily active users", "", "grid", "timeseries", "last_24hrs", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, b.ContactEntity.ID, NoEntityID, "Chrun rate", "life_cycle_stage", "grid", "normal", "last_24hrs", chart.CalcCount)
	if err != nil {
		return err
	}
	err = b.ChartAdd(ctx, b.ContactEntity.ID, NoEntityID, "New Customer", "became_a_customer_date", "grid", "normal", "last_24hrs", chart.CalcCount)
	if err != nil {
		return err
	}

	err = b.ChartAdd(ctx, b.ProjectEntity.ID, b.CompanyEntity.ID, "Delayed Accounts", "end_time", "grid", "normal", "last_24hrs", chart.CalcCount)
	if err != nil {
		return err
	}
	return nil
}
