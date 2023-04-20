package handlers_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/cmd/relay-api/internal/handlers"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestSetIncident(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	t.Log("Change the incident status")
	{

		t.Log("\tIncident status should be changed")
		{
			err := handlers.SetIncidentStatus(tests.Context(), schema.SeedAccountID, "69d2af4e-61a5-4239-988c-d5e1bd7bc4e7", "49a53e08-033a-40a2-912f-171a940e66e6", "acknoweldged", db, nil)
			if err != nil {
				t.Fatalf("\tShould be able to update incidents - %s", err)
			}
		}

	}
}
