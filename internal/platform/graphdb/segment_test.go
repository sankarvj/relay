package graphdb_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	accountID       = "2c247443-b257-4b06-ba99-493cf9d83ce7"
	contactEntityID = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	taskEntityID1   = "8d9c4f94-890b-484c-8189-91c3d7e8e50c"
	dealEntityID    = "109c4f94-890b-484c-8189-91c3d7e8e50c"
	contactItemID   = "12345"
	taskItemID1     = "54321"
	taskItemID2     = "99999"
	dealItemID      = "26436"
	fieldID1        = "4d247443-b257-4b06-ba99-493cf9d83ce7"
	taskRefFieldID  = "5d247443-b257-4b06-ba99-493cf9d83ce7"
	dealRefFieldID  = "33333343-b257-4b06-ba99-493cf9d83ce7"
	Name1           = "Panchavan Paari Venthan"
	Name2           = "Kosakshi Pasapughaz"
	colors          = []string{"blue", "yellow"}
	ref             = []string{taskItemID1}
	sref            = []string{taskItemID2}

	//gbp0
	taskProperties1 = map[string]interface{}{
		"name":  "Task1",
		"score": 100,
	}
	taskProperties2 = map[string]interface{}{
		"name":  "Task2",
		"score": 1000,
	}
	taskEntityFields = []graphdb.Field{
		graphdb.Field{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		graphdb.Field{
			Key:      "score",
			DataType: graphdb.TypeNumber,
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
		graphdb.Field{
			Key:      "id",
			DataType: graphdb.TypeString,
		},
		graphdb.Field{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		graphdb.Field{
			Key:      "age",
			DataType: graphdb.TypeNumber,
		},
		graphdb.Field{
			Key:      fieldID1,
			DataType: graphdb.TypeList,
			Field: &graphdb.Field{
				Key:      "element",
				DataType: graphdb.TypeString,
			},
		},
		graphdb.Field{
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
		graphdb.Field{
			Key:      "name",
			DataType: graphdb.TypeString,
			Value:    Name2,
		},
		graphdb.Field{
			Key:          fieldID1,
			DataType:     graphdb.TypeList,
			Value:        []string{"white", "blue"},
			UnlinkOffset: 2, // this will remove blue and add white. Yellow will persist
			Field: &graphdb.Field{
				Key:      "element",
				DataType: graphdb.TypeString,
			},
		},
		graphdb.Field{
			Key:          taskRefFieldID,
			DataType:     graphdb.TypeReference,
			RefID:        taskEntityID1,
			UnlinkOffset: 2, // this will remove old task and set a new task relation
			Value:        []string{taskItemID2, taskItemID1},
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
		dealRefFieldID: []string{contactItemID},
	}
	dealEntityFields = []graphdb.Field{
		graphdb.Field{
			Key:      "name",
			DataType: graphdb.TypeString,
		},
		graphdb.Field{
			Key:      "amount",
			DataType: graphdb.TypeNumber,
		},
		graphdb.Field{
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
		graphdb.Field{
			Expression: "<",
			Key:        "age",
			DataType:   graphdb.TypeNumber,
			Value:      "50",
		},
		graphdb.Field{
			Key:      fieldID1,
			DataType: graphdb.TypeList,
			Value:    []string{"yellow"},
			Field: &graphdb.Field{
				Expression: "=",
				Key:        "element",
				DataType:   graphdb.TypeString,
			},
		},
		graphdb.Field{
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
		graphdb.Field{
			Value:    []string{""},
			RefID:    dealEntityID,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{
				Expression: ">",
				Key:        "amount",
				DataType:   graphdb.TypeNumber,
				Value:      998,
				Aggr:       "SUM",
			},
		},
		graphdb.Field{
			Value:    []string{""},
			RefID:    dealEntityID,
			DataType: graphdb.TypeReference,
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
			_, err := graphdb.GetResult(residPool, gSegment)
			if err != nil {
				t.Fatalf("\t%s should fetch with segmentation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should fetch with segmentation", tests.Success)
			//case2
			// if n.GetProperty("name") != Name2 {
			// 	t.Fatalf("\t%s should fetch the node with %s - %s", tests.Failed, Name2, err)
			// }
			// t.Logf("\t%s should fetch the node with %s", tests.Success, Name2)
		}
	}

}
