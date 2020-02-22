package entity_test

import (
	"testing"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestDataEntity(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	t.Log("Given the need to create an data entity.")
	{
		t.Log("\tWhen adding the data entity")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			field := entity.Field{}
			fields := []entity.Field{field}

			ne := entity.NewEntity{
				TeamID: string(3),
				Name:   "",
				Fields: fields,
			}
			entity.Create(ctx, db, ne, now)
		}
	}

}
