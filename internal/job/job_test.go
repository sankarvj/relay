package job

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestRelationshipMap(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)

	t.Log("Fetch the relationship map")
	{
		t.Log("\twhen fetching the relationship map")
		{
			relationMap(tests.Context(), db, schema.SeedAccountID, "17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44")
		}
	}
}
