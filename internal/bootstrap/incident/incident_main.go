package incident

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/dashboard"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)
	err := loadPriorityType(ctx, b)
	if err != nil {
		fmt.Println("\tINCIDENT:BOOT Error  - Incident Type/Priority Created")
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Type/Priority/Category Created")

	// add entity - incident
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityIncidents, "Incidents", entity.CategoryData, entity.StateTeamLevel, false, true, false, IncidentFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.IncidentPriorityEntity.ID, b.IncidentPriorityEntity.Key("name"), b.IncidentTypeEntity.ID, b.IncidentTypeEntity.Key("name"), b.IncidentCategoryEntity.ID, b.IncidentCategoryEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Entity Created")

	// add entity - alerts
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAlerts, "Alerts", entity.CategoryData, entity.StateTeamLevel, false, true, false, AlertFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.IncidentPriorityEntity.ID, b.IncidentPriorityEntity.Key("name"), b.IncidentTypeEntity.ID, b.IncidentTypeEntity.Key("name"), b.IncidentCategoryEntity.ID, b.IncidentCategoryEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Alert Entity Created")

	// add entity - bugs
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityBugs, "Bugs", entity.CategoryData, entity.StateTeamLevel, false, true, false, AlertFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.IncidentPriorityEntity.ID, b.IncidentPriorityEntity.Key("name"), b.IncidentTypeEntity.ID, b.IncidentTypeEntity.Key("name"), b.IncidentCategoryEntity.ID, b.IncidentCategoryEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Bug Entity Created")

	// add entity - tasks
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTask, "Tasks", entity.CategoryTask, entity.StateTeamLevel, false, true, false, TaskFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}

	fmt.Println("\tINCIDENT:BOOT Tasks Entity Created")

	// add entity - approvals
	fmt.Println("\ttINCIDENT:BOOT Approvals Entity Started")
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityApprovals, "Approvals", entity.CategoryApprovals, entity.StateTeamLevel, false, false, false, forms.ApprovalsFields(b.ApprovalStatusEntity.ID, b.ApprovalStatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\ttINCIDENT:BOOT Approvals Entity Created")

	// add entity - notes
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNote, "Notes", entity.CategoryNotes, entity.StateTeamLevel, false, false, false, NoteFields())
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Notes Entity Created")

	return nil
}

func AddWorkflows(ctx context.Context, b *base.Base) error {
	err := addAutomation(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:SAMPLES Automations Created")

	incidentEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityIncidents)
	if err != nil {
		return err
	}

	err = b.AddSegments(ctx, incidentEntity.ID)
	if err != nil {
		return err
	}

	fmt.Println("\tINCIDENT:SAMPLES Sample Segments Created For Incidents")
	return nil
}

func AddSamples(ctx context.Context, b *base.Base) error {

	emailsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmails)
	if err != nil {
		return err
	}

	incidentEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityIncidents)
	if err != nil {
		return err
	}

	alertsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAlerts)
	if err != nil {
		return err
	}

	bugsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityBugs)
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

	err = addAssociations(ctx, b, incidentEntity, alertsEntity, bugsEntity, emailsEntity, streamEntity, taskEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	err = addStreams(ctx, b, streamEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:SAMPLES Streams Item Created")

	err = addDashboards(ctx, b, incidentEntity, taskEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:SAMPLES Charts Created")

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, inciEid, alertEid, bugsEid, emailEid, streamEID, taskEID, approvalsEID entity.Entity) error {

	var err error
	//task approvals association
	_, err = b.AssociationAdd(ctx, taskEID.ID, approvalsEID.ID)
	if err != nil {
		log.Println("ignoring error here. because the contraint might fail if it is alreay added in CRP")
		return err
	}
	//incident task association
	_, err = b.AssociationAdd(ctx, inciEid.ID, taskEID.ID)
	if err != nil {
		return err
	}
	//alert task association
	_, err = b.AssociationAdd(ctx, alertEid.ID, taskEID.ID)
	if err != nil {
		return err
	}
	//bugs task association
	_, err = b.AssociationAdd(ctx, bugsEid.ID, taskEID.ID)
	if err != nil {
		return err
	}

	//ASSOCIATE STREAMS
	//incident/alert/bugs stream association
	_, err = b.AssociationAdd(ctx, inciEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	_, err = b.AssociationAdd(ctx, alertEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	_, err = b.AssociationAdd(ctx, bugsEid.ID, streamEID.ID)
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

func addDashboards(ctx context.Context, b *base.Base, incidentEntity, taskEntity, approvalsEntity entity.Entity) error {

	homeDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, entity.NoEntityID, "Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addHomeCharts(ctx, b, homeDashID, incidentEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:Dashboard Overview Created")
	incDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, incidentEntity.ID, "Incident Overview").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addIncidentCharts(ctx, b, incDashID, incidentEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:Dashboard Incident Overview Created")
	myDashID, err := dashboard.BuildNewDashboard(b.AccountID, b.TeamID, b.UserID, b.NotificationEntity.ID, "My Dashboard").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = addMyCharts(ctx, b, myDashID, incidentEntity, taskEntity, approvalsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:Dashboard My Dashboard Created")

	return nil
}

func addHomeCharts(ctx context.Context, b *base.Base, dashboardID string, incidentEntity entity.Entity) error {
	//charts for home dashboard
	var err error

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "incidents_category", "Incidents category", "category", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "incidents_type", "Incidents type", "type", chart.TypeBar).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		return err
	}

	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "new_incidents", "Incidents rate", "status", chart.TypeGrid).AddDateField("system_created_at").SetDurationLast24hrs().SetCalcRate().SetGrpLogicID().SetIcon("chrun.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "sum_incidents", "New Incidents", "", chart.TypeGrid).AddDateField("system_created_at").SetDurationLast24hrs().SetCalcSum().SetIcon("new_customer.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}
	exp1 := fmt.Sprintf("{{%s.%s}} bf {%s}", incidentEntity.ID, incidentEntity.Key("end_time"), "now")
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "delayed", "Delayed", "end_time", chart.TypeGrid).AddExp(exp1).AddSource(b.CompanyEntity.ID).SetDurationLast24hrs().SetCalcSum().SetGrpLogicParent().SetIcon("delayed.svg").Add(ctx, b.DB)
	if err != nil {
		return err
	}

	return nil
}

func addIncidentCharts(ctx context.Context, b *base.Base, dashboardID string, incidentEntity, taskEntity entity.Entity) error {
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, taskEntity.ID, "tasks", "Tasks", "status", chart.TypePie).SetGrpLogicID().SetDurationAllTime().Add(ctx, b.DB)
	if err != nil {
		log.Println("error in tasks status")
		return err
	}

	return nil
}

func addMyCharts(ctx context.Context, b *base.Base, dashboardID string, incidentEntity, taskEntity, approvalsEntity entity.Entity) error {
	//charts for notifications me
	paOnMeExp := fmt.Sprintf("{{%s.%s}} in {%s,%s} && {{%s.%s}} in {{me}}", approvalsEntity.ID, approvalsEntity.Key("status"), b.ApprovalStatusWaiting.ID, b.ApprovalStatusChangeRequested.ID, approvalsEntity.ID, approvalsEntity.Key("assignees"))
	err := chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, approvalsEntity.ID, "my_pending_approvals", "My Pending Approvals", "", chart.TypeCard).AddExp(paOnMeExp).SetDurationAllTime().SetGrpLogicID().SetIcon("pending-approvals.svg").Add(ctx, b.DB)
	if err != nil {
		log.Println("error in my_pending_approvals")
		return err
	}
	overdueOnMeEXP := fmt.Sprintf("{{%s.%s}} !eq {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", taskEntity.ID, taskEntity.Key("status"), b.StatusItemClosed.ID, taskEntity.ID, taskEntity.Key("due_by"), "now", taskEntity.ID, taskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, taskEntity.ID, "my_overdue_tasks", "My Overdue Tasks", "", chart.TypeCard).AddExp(overdueOnMeEXP).SetDurationAllTime().SetIcon("overdue-tasks.svg").Add(ctx, b.DB)
	if err != nil {
		log.Println("error in my_overdue_tasks")
		return err
	}
	openOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", taskEntity.ID, taskEntity.Key("status"), b.StatusItemOpened.ID, taskEntity.ID, taskEntity.Key("assignees"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, taskEntity.ID, "my_open_tasks", "My Open Tasks", "", chart.TypeCard).AddExp(openOnMeEXP).SetDurationAllTime().SetIcon("open-tasks.svg").Add(ctx, b.DB)
	if err != nil {
		log.Println("error in my_open_tasks")
		return err
	}
	overdueProjOnMeEXP := fmt.Sprintf("{{%s.%s}} !in {%s} && {{%s.%s}} bf {%s} && {{%s.%s}} in {{me}}", incidentEntity.ID, incidentEntity.Key("status"), b.StatusItemClosed.ID, incidentEntity.ID, incidentEntity.Key("end_time"), "now", incidentEntity.ID, incidentEntity.Key("owner"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "my_overdue_incidents", "My Overdue Incidents", "", chart.TypeCard).AddExp(overdueProjOnMeEXP).SetDurationAllTime().SetIcon("overdue-projects.svg").Add(ctx, b.DB)
	if err != nil {
		log.Println("error in my_overdue_incidents")
		return err
	}
	openProjOnMeEXP := fmt.Sprintf("{{%s.%s}} in {%s} && {{%s.%s}} in {{me}}", incidentEntity.ID, incidentEntity.Key("status"), b.StatusItemOpened.ID, incidentEntity.ID, incidentEntity.Key("owner"))
	err = chart.BuildNewChart(b.AccountID, b.TeamID, dashboardID, incidentEntity.ID, "my_open_incidents", "My Open Incidents", "", chart.TypeCard).AddExp(openProjOnMeEXP).SetDurationAllTime().SetIcon("open-projects.svg").Add(ctx, b.DB)
	if err != nil {
		log.Println("error in my_open_incidents")
		return err
	}

	return nil
}

func loadPriorityType(ctx context.Context, b *base.Base) error {
	var err error
	// add entity - Type
	b.IncidentTypeEntity, err = b.EntityAdd(ctx, uuid.New().String(), "incident_type", "Incident Type", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, TypeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Type Entity Created")
	// add type item - warning
	_, err = b.ItemAdd(ctx, b.IncidentTypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentTypeEntity.Key("icon"), b.IncidentTypeEntity.Key("name"), b.IncidentTypeEntity.Key("color"), "🚧", "Warning", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add type item - error
	_, err = b.ItemAdd(ctx, b.IncidentTypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentTypeEntity.Key("icon"), b.IncidentTypeEntity.Key("name"), b.IncidentTypeEntity.Key("color"), "🛑", "Error", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add type item - bug
	_, err = b.ItemAdd(ctx, b.IncidentTypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentTypeEntity.Key("icon"), b.IncidentTypeEntity.Key("name"), b.IncidentTypeEntity.Key("color"), "🐞", "Bug", "#BD4291"), nil)
	if err != nil {
		return err
	}
	// add type item - service down
	_, err = b.ItemAdd(ctx, b.IncidentTypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentTypeEntity.Key("icon"), b.IncidentTypeEntity.Key("name"), b.IncidentTypeEntity.Key("color"), "💣", "Service Down", "#F94C66"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Types Created")

	// add entity - Category
	b.IncidentCategoryEntity, err = b.EntityAdd(ctx, uuid.New().String(), "incident_category", "Incident Category", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, TypeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Category Entity Created")
	// add type item - anomaly
	_, err = b.ItemAdd(ctx, b.IncidentCategoryEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentCategoryEntity.Key("icon"), b.IncidentCategoryEntity.Key("name"), b.IncidentCategoryEntity.Key("color"), "🎲", "Anomaly", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add type item - frequent
	_, err = b.ItemAdd(ctx, b.IncidentCategoryEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentCategoryEntity.Key("icon"), b.IncidentCategoryEntity.Key("name"), b.IncidentCategoryEntity.Key("color"), "⚡", "Frequent", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add type item - rare
	_, err = b.ItemAdd(ctx, b.IncidentCategoryEntity.ID, uuid.New().String(), b.UserID, TypeVals(b.IncidentCategoryEntity.Key("icon"), b.IncidentCategoryEntity.Key("name"), b.IncidentCategoryEntity.Key("color"), "🌋", "Rare", "#BD4291"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Categories Created")

	// add entity - agile priority
	b.IncidentPriorityEntity, err = b.EntityAdd(ctx, uuid.New().String(), "incident_priority", "Incident Priority", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, PriorityFields())
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident Priority Entity Created")
	// add priority item - low
	_, err = b.ItemAdd(ctx, b.IncidentPriorityEntity.ID, uuid.New().String(), b.UserID, PriorityVals(b.IncidentPriorityEntity.Key("name"), b.IncidentPriorityEntity.Key("color"), "Low", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add priority item - normal
	_, err = b.ItemAdd(ctx, b.IncidentPriorityEntity.ID, uuid.New().String(), b.UserID, PriorityVals(b.IncidentPriorityEntity.Key("name"), b.IncidentPriorityEntity.Key("color"), "Normal", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add priority item - high
	_, err = b.ItemAdd(ctx, b.IncidentPriorityEntity.ID, uuid.New().String(), b.UserID, PriorityVals(b.IncidentPriorityEntity.Key("name"), b.IncidentPriorityEntity.Key("color"), "High", "#BD4291"), nil)
	if err != nil {
		return err
	}
	// add priority item - critical
	_, err = b.ItemAdd(ctx, b.IncidentPriorityEntity.ID, uuid.New().String(), b.UserID, PriorityVals(b.IncidentPriorityEntity.Key("name"), b.IncidentPriorityEntity.Key("color"), "Critical", "#F94C66"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tINCIDENT:BOOT Incident priorities Created")
	return nil
}
