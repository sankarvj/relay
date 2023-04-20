package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/twilio/twilio-go/twiml"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

type Twilio struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

// Action gathers input from the outbound call
func (twil *Twilio) Action(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {

	body, err := getBody(r.Body)
	if err != nil {
		return errors.Wrap(err, "unable to decode body when the reader is passed")
	}

	accountID := params["account_id"]
	accountToken := params["account_token"]
	entityID := params["entity_id"]
	itemID := params["item_id"]
	fmt.Println("validate accountToken ::", accountToken)

	bodyParams, _ := url.ParseQuery(string(body))

	ivrInput := bodyParams.Get("Digits")
	if ivrInput != "" {
		if ivrInput == "1" {
			err = SetIncidentStatus(ctx, accountID, entityID, itemID, "acknowledged", twil.db, twil.sdb)
		} else {
			err = SetIncidentStatus(ctx, accountID, entityID, itemID, "triggered", twil.db, twil.sdb)
		}
	}

	statusCallbackInput := bodyParams.Get("CallStatus")
	if statusCallbackInput != "" && statusCallbackInput == "ringing" {
		// set status as triggered
		err = SetIncidentStatus(ctx, accountID, entityID, itemID, "triggered", twil.db, twil.sdb)
	}

	respondBack := &twiml.VoiceSay{
		Message:  "Thanks",
		Voice:    "woman",
		Language: "en-US",
	}

	if err != nil {
		respondBack = &twiml.VoiceSay{
			Message:  "Can't change the incident status. Call support",
			Voice:    "woman",
			Language: "en-US",
		}
	}

	verbList := []twiml.Element{respondBack}
	twimlResult, err := twiml.Voice(verbList)
	if err != nil {
		return err
	}
	return web.RespondXML(ctx, w, twimlResult, http.StatusOK)
}

func SetIncidentStatus(ctx context.Context, accountID, entityID, itemID, status string, db *sqlx.DB, sdb *database.SecDB) error {
	e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return err
	}

	it, err := item.Retrieve(ctx, accountID, entityID, itemID, db)
	if err != nil {
		return err
	}

	incidentFields := it.Fields()

	ff := e.FlowField()
	if ff != nil {
		flowIDs := it.Fields()[ff.Key]
		if flowIDs != nil { // make sure to create the incident with pipeline flow
			flowID := flowIDs.([]interface{})[0]
			viewModelNodes, err := nodeStages(ctx, accountID, flowID.(string), db)
			if err != nil {
				return err
			}

			nf := e.NodeField()
			if nf != nil {
				nodeIDs := it.Fields()[nf.Key]

				if nodeIDs != nil && len(nodeIDs.([]interface{})) > 0 {
					nodeID := nodeIDs.([]interface{})[0]
					for _, vmn := range viewModelNodes {
						if nodeID == vmn.ID {
							if vmn.Exp == "open" || vmn.Exp == "triggered" || vmn.Exp == "acknowledged" {
								for _, vmn := range viewModelNodes {
									if vmn.Expression == status {
										return updateField(ctx, accountID, entityID, itemID, incidentFields, nf.Key, vmn.ID, db)
									}
								}
							} else {
								log.Println("cominghere...")
							}
						}
					}
				} else {
					for _, vmn := range viewModelNodes {
						if vmn.Expression == status {
							return updateField(ctx, accountID, entityID, itemID, incidentFields, nf.Key, vmn.ID, db)
						}
					}
				}
			}

		}
	}

	return nil
}

func updateField(ctx context.Context, accountID, entityID, itemID string, fields map[string]interface{}, nodeFieldKey, nodeFieldValue string, db *sqlx.DB) error {
	fields[nodeFieldKey] = []interface{}{nodeFieldValue}
	it, err := item.UpdateFields(ctx, db, accountID, entityID, itemID, fields)
	if err != nil {
		return errors.Wrapf(err, "node value updated in incidents: %+v", &it)
	}
	return nil
}
