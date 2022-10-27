package base

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func bootstrapStatusEntity(ctx context.Context, b *Base) error {
	var err error
	// add entity - status
	fields := forms.StatusFields()
	b.StatusEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityStatus, "Status", entity.CategoryChildUnit, entity.StateAccountLevel, false, false, true, fields)
	if err != nil {
		return err
	}

	// add status item - in-progess
	b.StatusItemOpened, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, forms.StatusVals(b.StatusEntity, entity.FuExpNone, "In-progress", "#FFEF82"), nil)
	if err != nil {
		return err
	}
	// add status item - completed
	b.StatusItemClosed, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, forms.StatusVals(b.StatusEntity, entity.FuExpDone, "Completed", "#B4E197"), nil)
	if err != nil {
		return err
	}
	// add status item - blocked
	b.StatusItemOverDue, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, forms.StatusVals(b.StatusEntity, entity.FuExpNeg, "Blocked", "#FF8C8C"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT Status Entity With It's Three Statuses Items Created")

	return nil
}

func bootstrapApprovalStatusEntity(ctx context.Context, b *Base) error {
	var err error
	// add entity - approval status
	fields := forms.ApprovalStatusFields()
	b.ApprovalStatusEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityApprovalStatus, "Approval Status", entity.CategoryChildUnit, entity.StateAccountLevel, false, false, true, fields)
	if err != nil {
		return err
	}

	// add status item - waiting
	b.ApprovalStatusWaiting, err = b.ItemAdd(ctx, b.ApprovalStatusEntity.ID, uuid.New().String(), b.UserID, forms.ApprovalStatusVals(b.ApprovalStatusEntity, entity.FuExpNone, "Waiting for approval", "waiting_for_approval", "#79DAE8"), nil)
	if err != nil {
		return err
	}
	// add status item - change requested
	b.ApprovalStatusChangeRequested, err = b.ItemAdd(ctx, b.ApprovalStatusEntity.ID, uuid.New().String(), b.UserID, forms.ApprovalStatusVals(b.ApprovalStatusEntity, entity.FuExpNeg, "Change requested", "change_requested", "#FFEF82"), nil)
	if err != nil {
		return err
	}
	// add status item - approved
	b.ApprovalStatusApproved, err = b.ItemAdd(ctx, b.ApprovalStatusEntity.ID, uuid.New().String(), b.UserID, forms.ApprovalStatusVals(b.ApprovalStatusEntity, entity.FuExpDone, "Approved", "approved", "#B4E197"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT Approval Status Entity With It's Three Approval Statuses Items Created")

	return nil
}
