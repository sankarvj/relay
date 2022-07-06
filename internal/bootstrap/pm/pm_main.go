package pm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)

	// add entity - agile status
	agileStatusEntity, err := b.EntityAdd(ctx, uuid.New().String(), "agile_status", "Agile Status", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, AgileTaskStatusFields())
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Status Entity Created")
	// add status item - open
	_, err = b.ItemAdd(ctx, agileStatusEntity.ID, uuid.New().String(), b.UserID, AgileStatusVals(agileStatusEntity.Key("name"), agileStatusEntity.Key("color"), "Open", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add status item - in-progress
	_, err = b.ItemAdd(ctx, agileStatusEntity.ID, uuid.New().String(), b.UserID, AgileStatusVals(agileStatusEntity.Key("name"), agileStatusEntity.Key("color"), "In Progress", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add status item - review
	_, err = b.ItemAdd(ctx, agileStatusEntity.ID, uuid.New().String(), b.UserID, AgileStatusVals(agileStatusEntity.Key("name"), agileStatusEntity.Key("color"), "Review", "#BD4291"), nil)
	if err != nil {
		return err
	}
	// add status item - closed
	_, err = b.ItemAdd(ctx, agileStatusEntity.ID, uuid.New().String(), b.UserID, AgileStatusVals(agileStatusEntity.Key("name"), agileStatusEntity.Key("color"), "Closed", "#F94C66"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Status Items Created")

	// add entity - agile priority
	agilePriorityEntity, err := b.EntityAdd(ctx, uuid.New().String(), "priority", "Agile Priority", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, AgileTaskPriorityFields())
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Priority Entity Created")
	// add priority item - low
	_, err = b.ItemAdd(ctx, agilePriorityEntity.ID, uuid.New().String(), b.UserID, AgileTaskPriorityVals(agilePriorityEntity.Key("name"), agilePriorityEntity.Key("color"), "Low", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add priority item - normal
	_, err = b.ItemAdd(ctx, agilePriorityEntity.ID, uuid.New().String(), b.UserID, AgileTaskPriorityVals(agilePriorityEntity.Key("name"), agilePriorityEntity.Key("color"), "Normal", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add priority item - high
	_, err = b.ItemAdd(ctx, agilePriorityEntity.ID, uuid.New().String(), b.UserID, AgileTaskPriorityVals(agilePriorityEntity.Key("name"), agilePriorityEntity.Key("color"), "High", "#BD4291"), nil)
	if err != nil {
		return err
	}
	// add priority item - urgent
	_, err = b.ItemAdd(ctx, agilePriorityEntity.ID, uuid.New().String(), b.UserID, AgileTaskPriorityVals(agilePriorityEntity.Key("name"), agilePriorityEntity.Key("color"), "Urgent", "#F94C66"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Priority Items Created")

	// add entity - agile type
	agileTypeEntity, err := b.EntityAdd(ctx, uuid.New().String(), "agile_type", "Agile Type", entity.CategoryChildUnit, entity.StateTeamLevel, false, false, false, AgileTypeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Type Entity Created")
	// add type item - bug
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Bug", "#53BF9D"), nil)
	if err != nil {
		return err
	}
	// add type item - epic
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Epic", "#FFC54D"), nil)
	if err != nil {
		return err
	}
	// add type item - feature
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Feature", "#BD4291"), nil)
	if err != nil {
		return err
	}
	// add type item - project
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Project", "#F94C66"), nil)
	if err != nil {
		return err
	}
	// add type item - milestone
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Milestone", "#F94C66"), nil)
	if err != nil {
		return err
	}
	// add type item - initiative
	_, err = b.ItemAdd(ctx, agileTypeEntity.ID, uuid.New().String(), b.UserID, AgileTypeVals(agileTypeEntity.Key("name"), agileTypeEntity.Key("color"), "Initiative", "#F94C66"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Type Items Created")

	// add entity - agile task
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAgileTask, "Agile Tasks", entity.CategoryTask, entity.StateTeamLevel, false, true, false, AgileTaskFields(agileStatusEntity.ID, agileStatusEntity.Key("name"), agilePriorityEntity.ID, agilePriorityEntity.Key("name"), agileTypeEntity.ID, agileTypeEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("email")))
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Tasks Entity Created")

	// add entity - agile sub-task
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityAgileSubTask, "Sub Tasks", entity.CategoryTask, entity.StateTeamLevel, false, false, false, AgileTaskFields(agileStatusEntity.ID, agileStatusEntity.Key("name"), agilePriorityEntity.ID, agilePriorityEntity.Key("name"), agileTypeEntity.ID, agileTypeEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("email")))
	if err != nil {
		return err
	}
	fmt.Println("\tPM:BOOT Agile Sub Tasks Entity Created")

	return nil

}

func AddSamples(ctx context.Context, b *base.Base) error {
	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAgileTask)
	if err != nil {
		return err
	}

	subTaskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityAgileSubTask)
	if err != nil {
		return err
	}

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	err = AddAssociations(ctx, b, taskEntity, subTaskEntity)
	if err != nil {
		return err
	}

	err = AddAssociations(ctx, b, taskEntity, streamEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tPM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")
	return nil
}

func AddAssociations(ctx context.Context, b *base.Base, taskEid, streamEID entity.Entity) error {
	//task stream association
	_, err := b.AssociationAdd(ctx, taskEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	return nil
}
