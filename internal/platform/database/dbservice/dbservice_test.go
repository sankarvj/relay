package dbservice_test

import (
	"fmt"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
)

var (
	conditionFields = []graphdb.Field{
		{
			Expression: "<",
			Key:        "age",
			DataType:   graphdb.TypeNumber,
			Value:      "50",
		},
		{
			Key:      "fieldID1",
			DataType: graphdb.TypeList,
			Value:    []interface{}{"yellow"},
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "id",
				DataType:   graphdb.TypeString,
			},
		},
		{
			Key:      "taskRefFieldID",
			Value:    []interface{}{"23344232323"},
			RefID:    "taskEntityID1",
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "id",
				DataType:   graphdb.TypeString,
			},
		},
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      []string{"1234"},
		},
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      []string{"2345"},
		},
	}
)

func TestWhString(t *testing.T) {
	// _, teardown := tests.NewUnit(t)
	// defer teardown()

	//tests.SeedData(t, db)

	t.Log("Given the need to where string formation in the psql JsonB")
	{
		t.Log("\tform where clause string")
		{
			wh := dbservice.WhBuilder(conditionFields)
			fmt.Println("wh string:: ", wh)
		}

	}
}
