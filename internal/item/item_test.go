package item_test

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	accountID       = "2c247443-b257-4b06-ba99-493cf9d83ce7"
	contactEntityID = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	taskEntityID    = "8d9c4f94-890b-484c-8189-91c3d7e8e50c"
	dealEntityID    = "109c4f94-890b-484c-8189-91c3d7e8e50c"
	contactItemID   = "12345"
	taskItemID      = "54321"
	dealItemID      = "26436"
	fieldID1        = "4d247443-b257-4b06-ba99-493cf9d83ce7"
	taskRefFieldID  = "5d247443-b257-4b06-ba99-493cf9d83ce7"
	dealRefFieldID  = "33333343-b257-4b06-ba99-493cf9d83ce7"
	Name1           = "Panchavan Pari Venthan"
	Name2           = "Kosakshi Pasapughaz"
	colors          = []string{"blue", "yellow"}
	ref             = entity.RefMap(taskEntityID, taskItemID)

	//gbp1
	contactProperties = map[string]interface{}{
		"id":           contactItemID,
		"name":         Name1,
		"age":          32,
		"male":         true,
		fieldID1:       colors,
		taskRefFieldID: ref,
	}
	// updated contact fields
	updatedFields = []entity.Field{
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
			Value:    Name2,
		},
	}
	//contact entity field skeleton
	contactEntityFields = []entity.Field{
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
			Key:      taskRefFieldID,
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Key:      "score",
				DataType: entity.TypeNumber,
			},
		},
	}

	//gbp2
	dealProperties = map[string]interface{}{
		"name":         "Deal1",
		"amount":       999,
		dealRefFieldID: entity.RefMap(contactEntityID, contactItemID),
	}
	dealEntityFields = []entity.Field{
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "amount",
			DataType: entity.TypeNumber,
		},
		entity.Field{
			Key:      dealRefFieldID,
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Key:      "name",
				DataType: entity.TypeString,
			},
		},
	}

	//gbp0
	taskProperties = map[string]interface{}{
		"name":  "Task1",
		"score": 100,
	}
	taskEntityFields = []entity.Field{
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "score",
			DataType: entity.TypeNumber,
		},
	}

	taskFields    = entity.FillFieldValues(taskEntityFields, taskProperties)
	gpb0          = item.BuildGNode(accountID, taskEntityID).MakeBaseGNode(taskItemID, taskFields)
	contactFields = entity.FillFieldValues(contactEntityFields, contactProperties)
	gpb1          = item.BuildGNode(accountID, contactEntityID).MakeBaseGNode(contactItemID, contactFields)
	dealFields    = entity.FillFieldValues(dealEntityFields, dealProperties)
	gpb2          = item.BuildGNode(accountID, dealEntityID).MakeBaseGNode(dealItemID, dealFields)
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
			Key:      taskRefFieldID,
			Value:    ref,
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Expression: "=",
				Key:        "score",
				DataType:   entity.TypeNumber,
				Value:      100,
			},
		},
		entity.Field{
			Value:    entity.RefMap(dealEntityID, ""),
			DataType: entity.TypeReference,
			Field: &entity.Field{
				Expression: ">",
				Key:        "amount",
				DataType:   entity.TypeNumber,
				Value:      998,
			},
		},
	}

	gSegment = item.BuildGNode(accountID, contactEntityID).SegmentBaseGNode(conditionFields)
)

func TestGraph(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	log.Printf("gpb1 %+v", gpb1)
	t.Log(" Given the need create nodes and edges")
	{

		t.Log("\twhen adding the task item to the graph")
		{
			_, err := item.UpsertNode(residPool, gpb0)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding the contact item to the graph with straight reference of task")
		{
			_, err := item.UpsertNode(residPool, gpb1)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding the deal item to the graph with reverse reference of contact")
		{
			_, err := item.UpsertNode(residPool, gpb2)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding a relation to the contact item to the graph")
		{
			_, err := item.UpsertEdge(residPool, gpb1)
			if err != nil {
				t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should make a relation", tests.Success)
		}

		t.Log("\twhen adding a relation to the deal item to the graph")
		{
			_, err := item.UpsertEdge(residPool, gpb2)
			if err != nil {
				t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should make a relation", tests.Success)
		}

		t.Log("\twhen fetching the created contact item from the graph")
		{
			n, err := item.GetNode(residPool, accountID, contactEntityID, contactItemID)
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

		t.Log("\twhen updating the existing contact item to the graph")
		{
			updateNameGbp := item.BuildGNode(accountID, contactEntityID).MakeBaseGNode(contactItemID, updatedFields)
			_, err := item.UpsertNode(residPool, updateNameGbp)
			if err != nil {
				t.Fatalf("\t%s should update the exisiting node(item) with %s - %s", tests.Failed, Name2, err)
			}
			t.Logf("\t%s should update the exisiting node(item) with %s", tests.Success, Name2)
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
