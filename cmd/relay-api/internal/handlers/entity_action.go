package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

func (e *Entity) Mark(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Update")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	var ve entity.ViewModelEntity
	if err := web.Decode(r, &ve); err != nil {
		return errors.Wrap(err, "")
	}

	err := entity.UpdateMarkers(ctx, e.db, e.sdb, accountID, entityID, ve.IsPublic, ve.IsCore, ve.IsShared, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Mark Entity: %+v", &ve)
	}

	return web.Respond(ctx, w, ve, http.StatusOK)
}

func (e *Entity) ShareTeam(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.EntityAction.ShareTeam")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	enty, err := entity.Retrieve(ctx, accountID, entityID, e.db, e.sdb)
	if err != nil {
		return err
	}

	var ve entity.Entity
	if err := web.Decode(r, &ve); err != nil {
		return errors.Wrap(err, "")
	}
	enty.SharedTeamIds = ve.SharedTeamIds

	err = entity.UpdateSharedTeam(ctx, e.db, e.sdb, accountID, entityID, enty.SharedTeamIds, time.Now())
	if err != nil {
		return errors.Wrapf(err, "unexpected error when sharing an entity with team: %s", params["shared_team_id"])
	}

	return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusOK)
}

func (e *Entity) RemoveTeam(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.EntityAction.RemoveTeam")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	enty, err := entity.Retrieve(ctx, accountID, entityID, e.db, e.sdb)
	if err != nil {
		return err
	}
	updatedTeamIds := make([]string, 0)
	for _, stID := range enty.SharedTeamIds {
		if stID != params["team_id"] {
			updatedTeamIds = append(updatedTeamIds, stID)
		}
	}
	enty.SharedTeamIds = updatedTeamIds

	err = entity.UpdateSharedTeam(ctx, e.db, e.sdb, accountID, enty.ID, enty.SharedTeamIds, time.Now())
	if err != nil {
		return errors.Wrapf(err, "unexpected error when removing an entity from shared team: %s", params["shared_team_id"])
	}

	return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusOK)
}

func (e *Entity) Associate(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.EntityAction.UpdateAssociation")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)

	var assReqBody AssociationReqBody
	if err := web.Decode(r, &assReqBody); err != nil {
		return errors.Wrap(err, "")
	}

	var errs []error

	for _, as := range assReqBody.AssociationReqs {
		var err error
		if as.Remove {
			err = relationship.Delete(ctx, e.db, accountID, as.RelationshipID)
		} else {
			_, err = relationship.Associate(ctx, e.db, accountID, entityID, as.DstEntityID)
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return web.Respond(ctx, w, errs, http.StatusInternalServerError)
	} else {
		return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
	}
}

func (e *Entity) Associations(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	entities, err := entity.TeamEntities(ctx, params["account_id"], params["team_id"], []int{}, e.db)
	if err != nil {
		return err
	}

	bonds, err := relationship.List(ctx, e.db, params["account_id"], params["team_id"], params["entity_id"], auth.God(ctx))
	if err != nil {
		return err
	}

	bondMap := make(map[string]string, 0)
	for _, b := range bonds {
		bondMap[b.EntityID] = b.RelationshipID
	}

	viewModelChildren := make([]ViewModelChildren, 0)
	for _, ent := range entities {
		if ent.TeamID == ent.AccountID || ent.ID == params["entity_id"] {
			continue
		}
		if relationshipID, ok := bondMap[ent.ID]; ok {
			viewModelChildren = append(viewModelChildren, createViewModelChildren(ent, relationshipID))
		} else {
			viewModelChildren = append(viewModelChildren, createViewModelChildren(ent, ""))
		}
	}

	return web.Respond(ctx, w, viewModelChildren, http.StatusOK)
}

func (e *Entity) ToggleAccess(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.EntityAction.ToggleAccess")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	if util.NotEmpty(entityID) {
		existingEntity, err := entity.Retrieve(ctx, accountID, entityID, e.db, e.sdb)
		if err != nil {
			return errors.Wrapf(err, "error retriving entity")
		}
		existingEntity.IsPublic = !existingEntity.IsPublic
		err = existingEntity.UpdatePublicAccess(ctx, e.db)
		if err != nil {
			return errors.Wrapf(err, "error when toggle is_public for entity")
		}
		return web.Respond(ctx, w, createViewModelEntity(existingEntity), http.StatusOK)
	}
	return web.Respond(ctx, w, "failure", http.StatusNotAcceptable)
}
