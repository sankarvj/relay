package graphdb_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	accountID        = "00247443-b257-4b06-ba99-493cf9d83ce7"
	contactEntityID  = "119c4f94-890b-484c-8189-91c3d7e8e50b"
	taskEntityID1    = "229c4f94-890b-484c-8189-91c3d7e8e50c"
	dealEntityID     = "339c4f94-890b-484c-8189-91c3d7e8e50c"
	statusEntityID   = "449c4f94-890b-484c-8189-91c3d7e8e50c"
	contactItemID    = "11110"
	taskItemID1      = "22220"
	taskItemID2      = "22221"
	dealItemID       = "33330"
	statusItemID     = "44440"
	fieldID1         = "1c247443-b257-4b06-ba99-493cf9d83ce7"
	taskRefFieldID   = "2t247443-b257-4b06-ba99-493cf9d83ce7"
	dealRefFieldID   = "3d333343-b257-4b06-ba99-493cf9d83ce7"
	statusRefFieldID = "4s333343-b257-4b06-ba99-493cf9d83ce7"
	Name1            = "Panchavan Paari Venthan"
	Name2            = "Kosakshi Pasapughaz"
	colors           = []interface{}{"blue", "yellow"}
	ref              = []interface{}{taskItemID1}
	sref             = []interface{}{taskItemID2}
	stref            = []interface{}{statusItemID}

	//gbp0
	taskProperties1 = map[string]interface{}{
		"name":           "Task1",
		"score":          100,
		statusRefFieldID: stref,
	}
	taskProperties2 = map[string]interface{}{
		"name":           "Task2",
		"score":          1000,
		statusRefFieldID: stref,
	}
	taskEntityFields = []graphdb.Field{
		{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		{
			Key:      "score",
			DataType: graphdb.TypeNumber,
		},
		{
			Key:      statusRefFieldID,
			DataType: graphdb.TypeReference,
			RefID:    statusEntityID,
			Field: &graphdb.Field{
				Key:      "id",
				DataType: graphdb.TypeNumber,
			},
		},
	}

	//gbp1
	contactProperties = map[string]interface{}{
		"id":           contactItemID,
		"name":         Name1,
		"age":          32,
		"male":         true,
		fieldID1:       colors,
		taskRefFieldID: ref,
	}

	//contact entity field skeleton
	contactEntityFields = []graphdb.Field{
		{
			Key:      "id",
			DataType: graphdb.TypeString,
		},
		{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		{
			Key:      "age",
			DataType: graphdb.TypeNumber,
		},
		{
			Key:      fieldID1,
			DataType: graphdb.TypeList,
			Field: &graphdb.Field{
				Key:      "element",
				DataType: graphdb.TypeString,
			},
		},
		{
			Key:      taskRefFieldID,
			DataType: graphdb.TypeReference,
			RefID:    taskEntityID1,
			Field: &graphdb.Field{
				Key:      "id",
				DataType: graphdb.TypeNumber,
			},
		},
	}

	// updated contact fields
	updatedFields = []graphdb.Field{
		{
			Key:      "name",
			DataType: graphdb.TypeString,
			Value:    Name2,
		},
		{
			Key:          fieldID1,
			DataType:     graphdb.TypeList,
			Value:        []interface{}{"white", "blue"},
			UnlinkOffset: 2, // this will remove blue and add white. Yellow will persist
			Field: &graphdb.Field{
				Key:      "element",
				DataType: graphdb.TypeString,
			},
		},
		{
			Key:          taskRefFieldID,
			DataType:     graphdb.TypeReference,
			RefID:        taskEntityID1,
			UnlinkOffset: 2, // this will remove old task and set a new task relation
			Value:        []interface{}{taskItemID2, taskItemID1},
			Field: &graphdb.Field{
				Key:      "id",
				DataType: graphdb.TypeNumber,
			},
		},
	}

	//gbp2
	dealProperties = map[string]interface{}{
		"name":         "Deal1",
		"amount":       999,
		dealRefFieldID: []interface{}{contactItemID},
	}
	dealEntityFields = []graphdb.Field{
		{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		{
			Key:      "amount",
			DataType: graphdb.TypeNumber,
		},
		{
			Key:      dealRefFieldID,
			DataType: graphdb.TypeReference,
			RefID:    contactEntityID,
			Field: &graphdb.Field{
				Key:      "id",
				DataType: graphdb.TypeString,
			},
		},
	}

	taskFields01  = graphdb.FillFieldValues(taskEntityFields, taskProperties1)
	gpb01         = graphdb.BuildGNode(accountID, taskEntityID1, false).MakeBaseGNode(taskItemID1, taskFields01)
	taskFields02  = graphdb.FillFieldValues(taskEntityFields, taskProperties2)
	gpb02         = graphdb.BuildGNode(accountID, taskEntityID1, false).MakeBaseGNode(taskItemID2, taskFields02)
	contactFields = graphdb.FillFieldValues(contactEntityFields, contactProperties)
	gpb1          = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode(contactItemID, contactFields)
	dealFields    = graphdb.FillFieldValues(dealEntityFields, dealProperties)
	gpb2          = graphdb.BuildGNode(accountID, dealEntityID, false).MakeBaseGNode(dealItemID, dealFields)
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
			Key:      fieldID1,
			DataType: graphdb.TypeList,
			Value:    []interface{}{"yellow"},
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "element",
				DataType:   graphdb.TypeString,
			},
		},
		{
			Key:      taskRefFieldID,
			Value:    sref,
			RefID:    taskEntityID1,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "score",
				DataType:   graphdb.TypeNumber,
				Value:      1000,
			},
		},
		{
			Key:      taskRefFieldID,
			Value:    []interface{}{""}, //same as above but no ID
			RefID:    taskEntityID1,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "score",
				DataType:   graphdb.TypeNumber,
				Value:      1000,
			},
		},
		{
			Value:     []interface{}{""},
			RefID:     dealEntityID,
			IsReverse: true,
			DataType:  graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: ">",
				Key:        "amount",
				DataType:   graphdb.TypeNumber,
				Value:      998,
				Aggr:       "SUM",
			},
		},
		{
			Value:     []interface{}{""},
			RefID:     dealEntityID,
			IsReverse: true,
			DataType:  graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: ">",
				Key:        "amount",
				DataType:   graphdb.TypeNumber,
				Value:      998,
				Aggr:       "MAX",
			},
		},
	}

	gSegment = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode("", conditionFields)

	conditionFieldsWithIN = []graphdb.Field{
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      []string{contactItemID},
		},
	}

	gSegment1 = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode("", conditionFieldsWithIN)

	conditionFieldsForCount = []graphdb.Field{
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      []string{contactItemID},
		},
		{
			Value:    []interface{}{""}, //this makes the relation between src and dst entity
			RefID:    taskEntityID1,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{ // this adds the condition to the relation over the task
				RefID:    statusRefFieldID,
				DataType: graphdb.TypeReference,
				Value:    []interface{}{""},
				Field: &graphdb.Field{
					Expression: "=",
					Key:        "id",
					DataType:   graphdb.TypeString,
					Value:      statusItemID, // status verb as done
				},
			},
		},
		// {
		// 	Value:    []interface{}{""}, //same as above but no ID
		// 	RefID:    taskEntityID1,
		// 	DataType: graphdb.TypeReference,
		// 	Field: &graphdb.Field{
		// 		Expression: "=",
		// 		Key:        "score",
		// 		DataType:   graphdb.TypeNumber,
		// 		Value:      1000,
		// 	},
		// },
	}

	gSegment2 = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode("", conditionFieldsForCount)
)

func TestGraph(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	//log.Printf("gpb1 %+v", gpb1)
	t.Log(" Given the need create nodes and edges")
	{
		t.Log("\twhen adding the task item 1 to the graph")
		{
			err := graphdb.UpsertNode(residPool, gpb01)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding the task item 2 to the graph")
		{
			err := graphdb.UpsertNode(residPool, gpb02)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding the contact item to the graph with straight reference of task")
		{
			err := graphdb.UpsertNode(residPool, gpb1)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		t.Log("\twhen adding the deal item to the graph with reverse reference of contact")
		{
			err := graphdb.UpsertNode(residPool, gpb2)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		// t.Log("\twhen adding a relation to the contact item to the graph")
		// {
		// 	_, err := graphdb.UpsertEdge(residPool, gpb1)
		// 	if err != nil {
		// 		t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
		// 	}
		// 	t.Logf("\t%s should make a relation", tests.Success)
		// }

		// t.Log("\twhen adding a relation to the deal item to the graph")
		// {
		// 	_, err := graphdb.UpsertEdge(residPool, gpb2)
		// 	if err != nil {
		// 		t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
		// 	}
		// 	t.Logf("\t%s should make a relation", tests.Success)
		// }

		t.Log("\twhen fetching the created contact item from the graph")
		{
			n, err := graphdb.GetNode(residPool, accountID, contactEntityID, contactItemID)
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
			updateNameGbp := graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode(contactItemID, updatedFields)
			err := graphdb.UpsertNode(residPool, updateNameGbp)
			if err != nil {
				t.Fatalf("\t%s should update the exisiting node(item) with %s - %s", tests.Failed, Name2, err)
			}
			t.Logf("\t%s should update the exisiting node(item) with %s", tests.Success, Name2)
		}

		t.Log("\twhen segmenting the updated item with relation to the graph")
		{
			_, err := graphdb.GetResult(residPool, gSegment, 0, "", "")
			if err != nil {
				t.Fatalf("\t%s should fetch with segmentation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should fetch with segmentation", tests.Success)
		}

		t.Log("\twhen querying with where clause as `IN ('itemID1')`")
		{
			_, err := graphdb.GetResult(residPool, gSegment1, 0, "", "")
			if err != nil {
				t.Fatalf("\t%s should query with IN clause - %s", tests.Failed, err)
			}
			t.Logf("\t%s should query with IN clause", tests.Success)
		}

		t.Log("\twhen querying the get count")
		{
			b, _ := json.Marshal(gSegment2)
			fmt.Println(string(b))
			_, err := graphdb.GetCount(residPool, gSegment2, true)
			if err != nil {
				t.Fatalf("\t%s should return count - %s", tests.Failed, err)
			}
			t.Logf("\t%s should return count", tests.Success)
		}
	}

}

var ()
