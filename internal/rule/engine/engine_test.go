package engine_test

import (
	"fmt"
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestEmailRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine to send email.")
	{
		t.Log("\twhen running a send email engine for the given contact - default case")
		{
			e1 := "17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44" //contacts-entity-id
			e2 := "447a9c03-ad0c-4543-9dcf-fce1a8fbceed" //mails-entity-id
			i1 := "9d9ab317-897d-4297-9818-088674ce497e" //contact-item-id
			i2 := "ef068193-426d-4155-b00b-378c8fcc82be" //mail-item-id (assuming the item has incorparated the contacts fields as vars in its body/subject)

			vars, _ := node.MapToJSONB(map[string]string{e1: i1}) // this will get populated only during the trigger
			acts, _ := node.MapToJSONB(map[string]string{e2: i2}) // this will get populated during the workflow creation

			node := node.Node{
				Expression: fmt.Sprintf("{{%s.uuid-00-fname}} eq {Vijay}", e1),
				Variables:  vars,
				Actuals:    acts,
				ActorID:    e2,
				Type:       node.Email,
			}

			_, err := engine.RunRuleEngine(tests.Context(), db, nil, node)
			if err != nil {
				t.Fatalf("\t%s\tshould send email : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tshould send the email", tests.Success)
		}
	}
}

func TestCreateItemRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine to create new item")
	{
		t.Log("\twhen running a create item engine - default case")
		{
			e1 := "17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44" //contacts-entity-id
			i1 := "9d9ab317-897d-4297-9818-088674ce497e" //contact-item-id
			e2 := "00000000-0000-0000-0000-000000000003" //task-entity-id
			i2 := "00000000-0000-0000-0000-000000000012" //task-item-id (the task blue-print which will be used to create other tasks)

			vars, _ := node.MapToJSONB(map[string]string{e1: i1})
			acts, _ := node.MapToJSONB(map[string]string{e2: i2})

			node := node.Node{
				//Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				AccountID: schema.SeedAccountID,
				Variables: vars,
				Actuals:   acts,
				ActorID:   e2,
				Type:      node.Push,
			}
			_, err := engine.RunRuleEngine(tests.Context(), db, nil, node)
			if err != nil {
				t.Fatalf("\t%s\tshould create item : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tshould create a item", tests.Success)
		}
	}
}

func TestUpdateRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine to update existing item")
	{
		t.Log("\twhen running update item engine - default case")
		{
			e1 := "00000000-0000-0000-0000-000000000002" //contacts-entity-id
			i1 := "00000000-0000-0000-0000-000000000010" //contact-item-id
			i2 := "00000000-0000-0000-0000-000000000011" //updatable-contact-id (Has blue-print of the values to be updated when triggered)

			vars, _ := node.MapToJSONB(map[string]string{e1: i1})
			acts, _ := node.MapToJSONB(map[string]string{e1: i2})

			node := node.Node{
				//Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				AccountID: schema.SeedAccountID,
				Variables: vars,
				Actuals:   acts,
				ActorID:   e1,
				Type:      node.Modify,
			}
			_, err := engine.RunRuleEngine(tests.Context(), db, nil, node)
			if err != nil {
				t.Fatalf("\t%s should update item : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should update a item", tests.Success)
		}
	}
}

func TestFlow(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine for a complete flow")
	{
		t.Log("\twhen running a flow 'The flow'")
		{
			f1 := "00000000-0000-0000-0000-000000000017" // flow-id
			flow, _ := flow.Retrieve(tests.Context(), f1, db)
			log.Printf("The flow %v", flow)
			nodes, _ := node.List(tests.Context(), f1, db)

			branchNodeMap := node.BranceNodeMap(nodes)
			rootNodes := node.ChildNodes(node.Root, branchNodeMap)
			log.Printf("The rootNodes %v", rootNodes)

			if len(rootNodes) > 0 {
				childNodes := node.ChildNodes(rootNodes[0].ID, branchNodeMap)
				log.Printf("The childNodes %v", childNodes)
			}

		}
	}
}

func TestTrigger(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine for a trigger")
	{
		t.Log("\tWhen updating the event mrr in contact1")
		{
			e1 := "00000000-0000-0000-0000-000000000002" //contact-entity-id
			i1 := "00000000-0000-0000-0000-000000000010" //contact-item-id
			i, _ := item.Retrieve(tests.Context(), e1, i1, db)
			oldItemFields := i.Fields()
			newItemFields := i.Fields()
			newItemFields["uuid-00-nps-score"] = 99
			item.UpdateFields(tests.Context(), db, e1, i1, newItemFields)
			// the above action will trigger this in the background thread
			flows, _ := flow.List(tests.Context(), []string{e1}, -1, db)
			dirtyFlows := flow.DirtyFlows(tests.Context(), flows, oldItemFields, newItemFields)
			err := flow.Trigger(tests.Context(), db, nil, i1, dirtyFlows)
			if err != nil {
				t.Fatalf("\t%s should flow without error : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should flow without error", tests.Success)
		}
	}
}

func TestDirectTrigger(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedPipelines(t, db)
	defer teardown()
	t.Log("Given the need to run the engine for a trigger")
	{
		t.Log("\twhen updating the event mrr in contact1")
		{
			e1 := "00000000-0000-0000-0000-000000000002"
			i1 := "00000000-0000-0000-0000-000000000010" //contact-item-id
			n2 := "00000000-0000-0000-0000-000000000025" //node-stage-2
			err := flow.DirectTrigger(tests.Context(), db, nil, n2, e1, i1)
			if err != nil {
				t.Fatalf("\t%s should flow without error : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should flow without error", tests.Success)

			afs, _ := flow.ActiveFlows(tests.Context(), []string{"00000000-0000-0000-0000-000000000023"}, db) //pipeline-id
			ans, _ := flow.ActiveNodes(tests.Context(), []string{"00000000-0000-0000-0000-000000000023"}, db) //pipeline-id
			log.Printf("afs >>>>>>>>>>>>>>>>>>>>>> %v", afs)
			log.Printf("ans >>>>>>>>>>>>>>>>>>>>>> %v", ans)

		}
	}
}

var (
	accountID       = "2c247443-b257-4b06-ba99-493cf9d83ce7"
	contactEntityID = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	contactItemID   = "12345"
	dealEntityID    = "109c4f94-890b-484c-8189-91c3d7e8e50c"
	dealRefFieldID  = "33333343-b257-4b06-ba99-493cf9d83ce7"
	dealItemID      = "26436"

	//gbp1
	contactEntityFields = []graphdb.Field{
		{
			Key:      "id",
			DataType: graphdb.TypeString,
		},
		{
			Key:      "age",
			DataType: graphdb.TypeNumber,
		},
	}

	contactProperties = map[string]interface{}{
		"id":  contactItemID,
		"age": 32,
	}

	//gbp2
	dealProperties = map[string]interface{}{
		"name":         "Deal1",
		"amount":       999,
		dealRefFieldID: []string{contactItemID},
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
			RefID:    contactEntityID,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{
				Key:      "name", //may be not needed
				DataType: graphdb.TypeString,
			},
		},
	}

	contactFields = graphdb.FillFieldValues(contactEntityFields, contactProperties)
	gpb1          = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode(contactItemID, contactFields)
	dealFields    = graphdb.FillFieldValues(dealEntityFields, dealProperties)
	gpb2          = graphdb.BuildGNode(accountID, dealEntityID, false).MakeBaseGNode(dealItemID, dealFields)
	//refer segment_test.go more complex conditions
	conditionFields = []graphdb.Field{
		{
			Expression: "<",
			Key:        "age",
			DataType:   graphdb.TypeNumber,
			Value:      "50",
		},
	}
	gSegment   = graphdb.BuildGNode(accountID, contactEntityID, false).MakeBaseGNode("", conditionFields)
	jsonB, _   = gSegment.JsonB()
	expression = fmt.Sprintf("<<%s>>", jsonB)
)

func TestQueryRuleRunner(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	t.Log("Given the need to run the engine to evaluate a query")
	{
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

		t.Log("\twhen segmenting the updated item with relation to the graph")
		{
			_, err := graphdb.GetResult(residPool, gSegment)
			if err != nil {
				t.Fatalf("\t%s should fetch the item - %s", tests.Failed, err)
			}
			t.Logf("\t%s should fetch the item", tests.Success)
		}

		t.Log("\twhen evaluating the query")
		{
			vars, _ := node.MapToJSONB(map[string]string{contactEntityID: contactItemID})
			node := node.Node{
				Expression: expression,
				AccountID:  accountID,
				Variables:  vars,
				Type:       node.Unknown,
			}
			_, err := engine.RunRuleEngine(tests.Context(), nil, residPool, node)
			if err != nil {
				t.Fatalf("\t%s should pass with out fail : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should pass with out fail", tests.Success)
		}
	}
}
