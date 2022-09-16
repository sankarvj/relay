package csm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
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
	projectEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityProjects, "Projects", entity.CategoryData, entity.StateTeamLevel, false, true, false, ProjectFields(b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("email"), b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Projects Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, true, false, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, projectEntity.ID, b.ContactEntity.Key("email"), b.ContactEntity.Key("first_name"), b.CompanyEntity.Key("name"), projectEntity.Key("project_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

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
	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityProjects)
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
