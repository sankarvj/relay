package item_test

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	accountID = "2c247443-b257-4b06-ba99-493cf9d83ce7"
	entityID1 = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	entityID2 = "8d9c4f94-890b-484c-8189-91c3d7e8e50c"
	itemID1   = "12345"
	itemID2   = "54321"
	fieldID1  = "4d247443-b257-4b06-ba99-493cf9d83ce7"
	fieldID2  = "5d247443-b257-4b06-ba99-493cf9d83ce7"
	Name1     = "Panchavan Pari Venthan"
	Name2     = "Kosakshi Pasapughaz"
	colors    = []string{"blue", "yellow"}
	ref       = entity.RefMap(entityID2, itemID2)

	//item
	properties = map[string]interface{}{
		"id":     itemID1,
		"name":   Name1,
		"age":    32,
		"male":   true,
		fieldID1: colors,
		fieldID2: ref,
	}
	// updated item
	updatedFields = []entity.Field{
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
			Value:    Name2,
		},
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
			Key:      fieldID1,
			DataType: entity.TypeList,
			Field: &entity.Field{
				Key:      "element",
				DataType: entity.TypeString,
			},
		},
		entity.Field{
			Key:      fieldID2,
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Key:      "score",
				DataType: entity.TypeNumber,
				Value:    100,
			},
		},
	}

	fields = entity.FillFieldValues(entityFields, properties)
	gpb    = item.BuildGNode(accountID, entityID1).MakeBaseGNode(itemID1, fields)
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
			Key:      fieldID1,
			DataType: entity.TypeList,
			Field: &entity.Field{
				Expression: "=",
				Key:        "element",
				DataType:   entity.TypeString,
				Value:      "yellow",
			},
		},
		entity.Field{
			Key:      fieldID2,
			Value:    ref,
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Expression: "=",
				Key:        "score",
				DataType:   entity.TypeNumber,
				Value:      100,
			},
		},
	}

	gSegment = item.BuildGNode(accountID, entityID1).SegmentBaseGNode(conditionFields)
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
			n, err := item.GetNode(residPool, accountID, entityID1, itemID1)
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
			updateNameGbp := item.BuildGNode(accountID, entityID1).MakeBaseGNode(itemID1, updatedFields)
			_, err := item.UpsertNode(residPool, updateNameGbp)
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
