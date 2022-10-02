package main

import (
	"encoding/json"
	"testing"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestEvent(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	sdb, teardown := tests.NewSecDbUnit(t)
	defer teardown()

	body := map[string]interface{}{
		"module":     "page_views",
		"identifier": "contacts:email:user@example.com",
		"event":      "User Registered",
		"count":      1,
		"icon":       "🔥",
		"notify":     true,
		"tags": map[string]interface{}{
			"email": "user@example.com",
			"uid":   "uid1234",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	reqBody := map[string]interface{}{
		"body": string(bodyBytes),
		"headers": map[string]interface{}{
			"authorization": "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W10sImV4cCI6MjI2ODU5MDU0OSwiaWF0IjoxNjYzNzkwNTQ5LCJzdWIiOiIzY2YxNzI2Ni0zNDczLTQwMDYtOTg0Zi05MzI1MTIyNjc4YjcifQ.dpZqxQLSroBYC3uutVTFuLuFTBvFpKPqGuQVlhMNyS-zonjko6foQ_9vxlbZr6Ax5D5tBD1EXxJb0RuU0kAQ3L-eFAbwbRHnUkaTODGqh1fwQnuxqvzUm9-tnvqYzz7jM8iyw1y21tEyynBB9Zx6cOLJGRN5FojIlHrVQ7P4KTxeBRsgzphl9bV5-5Ge-Ec8_-fKtuwGKDrNWPVjU6qbIDlgGNVdTcJcYDco5_KUcuUAvKmc3LrvmB8oQPZb1byjc0JnNbIxC2cRnCLwiphoceYsnSwrbxUlyGNNG1uutFqsqQZca6aNx62X6TqSmaw-6Kt7KGVqg9sv4y9JUP8oRQ",
		},
	}

	h := &EventsHandler{
		sdb: sdb,
		db:  db,
	}
	err := initializeAuth(h, "../../private.pem", "1", "config/dev/relay-70013-firebase-adminsdk-cfun3-58caec85f0.json", "config/dev/google-apps-client-secret.json", "RS256")
	if err != nil {
		t.Fatalf("\tShould be able to set authenticator - %s", err)
	}
	var pvE, stE, acE entity.NewEntity

	t.Log("Given the need to create or update an event.")
	{

		t.Log("\tPrepare page vist events entity")
		{
			pvE = entity.NewEntity{
				ID:          "00000000-0000-0000-0000-000000000111",
				DisplayName: "Page Views",
				Name:        "page_views",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      pageVisitFields(),
			}
			_, err := entity.Create(tests.Context(), db, pvE, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to create page visits entity - %s", err)
			}
		}

		t.Log("\tPrepare status events entity")
		{
			stE = entity.NewEntity{
				ID:          "00000000-0000-0000-0000-000000000112",
				DisplayName: "Status",
				Name:        "status",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      statusFields(),
			}
			_, err := entity.Create(tests.Context(), db, stE, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to create status entity - %s", err)
			}
		}

		t.Log("\tPrepare action events entity")
		{
			acE = entity.NewEntity{
				ID:          "00000000-0000-0000-0000-000000000113",
				DisplayName: "Actions",
				Name:        "actions",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      actionFields(),
			}
			_, err := entity.Create(tests.Context(), db, acE, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to create action entity - %s", err)
			}
		}

		t.Log("\tAdd an first page visit event")
		{
			_, err := h.handleEvent(tests.Context(), reqBody)
			//err := h.processEvent(tests.Context(), schema.SeedAccountID, pvE.ID, pageVisits(), "", "")
			if err != nil {
				t.Fatalf("\tShould be able to create an event - %s", err)
			}
		}

		t.Log("\tAdd an second page visit event")
		{
			_, err := h.handleEvent(tests.Context(), reqBody)
			//err := h.processEvent(tests.Context(), schema.SeedAccountID, pvE.ID, pageVisits(), "", "", db)
			if err != nil {
				t.Fatalf("\tShould be able to update an event - %s", err)
			}
		}

		t.Log("\tRetrieve page visit events")
		{
			events, err := item.List(tests.Context(), schema.SeedAccountID, pvE.ID, db)
			if err != nil || len(events) == 0 {
				t.Fatalf("\tShould be able to get atleast one event - %s", err)
			} else {
				t.Logf("\tAdded Events: %+v", events)
			}
		}
	}
}

func pageVisitFields() []entity.Field {
	nameField := entity.Field{
		Key:         "event",
		Name:        "event",
		DisplayName: "Event",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	countField := entity.Field{
		Key:         "count",
		Name:        "count",
		DisplayName: "Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyCalc: entity.MetaCalcSum},
	}

	return []entity.Field{nameField, countField}
}

func statusFields() []entity.Field {
	statusField := entity.Field{
		Key:         "status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{statusField}
}

func actionFields() []entity.Field {
	actionField := entity.Field{
		Key:         "action",
		Name:        "action",
		DisplayName: "Action",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	countField := entity.Field{
		Key:         "count",
		Name:        "count",
		DisplayName: "Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyCalc: entity.MetaCalcSum},
	}

	return []entity.Field{actionField, countField}
}

func pageVisits() map[string]interface{} {
	pageVisits := make(map[string]interface{}, 0)
	pageVisits["event"] = "user signup"
	pageVisits["count"] = 1
	return pageVisits
}

func status() map[string]interface{} {
	status := make(map[string]interface{}, 0)
	status["status"] = "up"
	return status
}
