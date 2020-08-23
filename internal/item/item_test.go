package item_test

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	accountID = "2d247443-b257-4b06-ba99-493cf9d83ce7"
	entityID  = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	itemID    = "12345"
	fieldID   = "4d247443-b257-4b06-ba99-493cf9d83ce7"
	Name1     = "Panchavan Pari Venthan"
	Name2     = "Kosakshi Pasapughaz"
	colors    = []string{"blue", "yellow"}

	//item
	properties = map[string]interface{}{
		"id":    itemID,
		"name":  Name1,
		"age":   32,
		"male":  true,
		fieldID: colors,
	}
	// updated item
	updatedProperties = map[string]interface{}{
		"name": Name2,
	}
	//entity field skeleton
	entityFields = []entity.Field{
		entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "age",
			DataType: entity.TypeNumber,
		},
		entity.Field{
			Key:      fieldID,
			DataType: entity.TypeList,
			Field: &entity.Field{
				Key:      "element",
				DataType: entity.TypeString,
			},
		},
	}

	fields = entity.FillFieldValues(entityFields, properties)
	gpb    = item.BuildGNode(accountID, entityID).MakeBaseGNode(itemID, fields)
)

var (
	conditionFields = []entity.Field{
		entity.Field{
			Expression: "<",
			Key:        "age",
			DataType:   entity.TypeNumber,
			Value:      "50",
		},
		entity.Field{
			Key:      fieldID,
			DataType: "L",
			Field: &entity.Field{
				Expression: "=",
				Key:        "element",
				DataType:   entity.TypeString,
				Value:      "yellow",
			},
		},
		entity.Field{
			Key:      fieldID, //replace here with the entityID
			DataType: "R",
			Field: &entity.Field{
				Expression: "=",
				Key:        "element", //fieldID of the entity
				DataType:   entity.TypeString,
				Value:      "yellow",
			},
		},
	}
	seg = entity.Segment{
		Match:           entity.MatchAll,
		FieldConditions: conditionFields,
	}

	gSegment = item.BuildGNode(accountID, entityID).SegmentBaseGNode(seg)
)

func TestGraph(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	log.Printf("gpb %+v", gpb)
	t.Log(" Given the need create nodes and edges")
	{
		t.Log("\twhen adding the new item to the graph")
		{
			_, err := item.UpsertNode(residPool, gpb)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen fetching the created item to the graph")
		{
			n, err := item.GetNode(residPool, accountID, entityID, itemID)
			if err != nil {
				t.Fatalf("\t%s should not throw any error during the fetch - %s", tests.Failed, err)
			}
			t.Logf("\t%s should not throw any error during the fetch", tests.Success)
			//case2
			if n.GetProperty("name") != Name1 {
				t.Fatalf("\t%s should fetch the node with %s - %s", tests.Failed, Name1, err)
			}
			t.Logf("\t%s should fetch the node with %s", tests.Success, Name1)
		}

		t.Log("\twhen updating the existing item to the graph")
		{
			gpb.Properties = updatedProperties
			_, err := item.UpsertNode(residPool, gpb)
			if err != nil {
				t.Fatalf("\t%s should update the exisiting node(item) with %s - %s", tests.Failed, Name2, err)
			}
			t.Logf("\t%s should update the exisiting node(item) with %s", tests.Success, Name2)
		}

		t.Log("\twhen adding a relation to the updated item to the graph")
		{
			_, err := item.UpsertEdge(residPool, gpb)
			if err != nil {
				t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should make a relation", tests.Success)
		}

		t.Log("\twhen fetching the updated item with relation to the graph")
		{
			_, err := item.GetResult(residPool, gSegment)
			if err != nil {
				t.Fatalf("\t%s should fetch with relation honda - %s", tests.Failed, err)
			}
			t.Logf("\t%s should fetch with relation honda", tests.Success)
			//case2
			// if n.GetProperty("name") != Name2 {
			// 	t.Fatalf("\t%s should fetch the node with %s - %s", tests.Failed, Name2, err)
			// }
			// t.Logf("\t%s should fetch the node with %s", tests.Success, Name2)
		}
	}

}

// var (
// 	complexConditions = []segment.Condition{
// 		segment.Condition{
// 			Operator: ">",
// 			Key:      "age",
// 			Type:     "N",
// 			Value:    "40",
// 		},
// 		segment.Condition{
// 			Operator: "<",
// 			Key:      "age",
// 			Type:     "N",
// 			Value:    "50",
// 		},
// 		segment.Condition{
// 			Operator: "=",
// 			Key:      "name",
// 			Type:     "S",
// 			Value:    "Siva",
// 		},
// 		segment.Condition{
// 			Operator: "=",
// 			EntityID: "colors",
// 			Key:      "element",
// 			Type:     "S",
// 			Value:    "blue",
// 			On:       segment.List,
// 		},
// 	}
// 	complexSeg = segment.Segment{
// 		Match:      segment.MatchAll,
// 		Conditions: complexConditions,
// 	}

// 	gSegmentCom = item.BuildGNode(accountID, entityID).SegmentBaseGNode(complexSeg)
// )

// func TestSegmentBaseGNode(t *testing.T) {
// 	residPool, teardown := tests.NewRedisUnit(t)
// 	defer teardown()
// 	t.Log(" Given the need to parse the segment into graph query")
// 	{
// 		t.Log("\twhen parsing AND conditions")
// 		{
// 			log.Printf("gbp1 ------> %+v", gSegmentCom)
// 			_, err := item.GetResult(residPool, gSegmentCom)
// 			if err != nil {
// 				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
// 			}
// 			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
// 		}
// 	}
// }
