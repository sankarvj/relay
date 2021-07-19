package engine_test

import (
	"fmt"
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
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
			contactEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedContactsEntityName)
			emailsEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, entity.FixedEntityEmails)
			contactItems, err := item.List(tests.Context(), contactEntity.ID, db)
			emailTemplateItems, _ := item.List(tests.Context(), emailsEntity.ID, db)

			vars, _ := node.MapToJSONB(map[string]string{contactEntity.ID: contactItems[0].ID})      // this will get populated only during the trigger
			acts, _ := node.MapToJSONB(map[string]string{emailsEntity.ID: emailTemplateItems[0].ID}) // this will get populated during the workflow creation

			node := node.Node{
				AccountID:  schema.SeedAccountID,
				Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", contactEntity.ID, schema.SeedFieldFNameKey),
				Variables:  vars,
				Actuals:    acts,
				ActorID:    emailsEntity.ID,
				Type:       node.Email,
			}

			eng := engine.Engine{
				Job: job.Job{},
			}

			_, err = eng.RunRuleEngine(tests.Context(), db, nil, node)
			if err != nil {
				t.Fatalf("\t%s\tshould send email : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tshould send the email", tests.Success)
		}
	}
}

func TestRuleRenderer(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	defer teardown()
	t.Log("Given the need to run the engine to render the template values.")
	{
		t.Log("\twhen running a engine for the given contact - default case")
		{
			contactEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedContactsEntityName)
			contactItems, _ := item.List(tests.Context(), contactEntity.ID, db)

			vars := map[string]interface{}{contactEntity.ID: contactItems[0].ID} // this will get populated only during the trigger
			exp := fmt.Sprintf("{{%s.%s}}", contactEntity.ID, schema.SeedFieldFNameKey)

			eng := engine.Engine{
				Job: job.Job{},
			}

			content := eng.RunExpRenderer(tests.Context(), db, schema.SeedAccountID, exp, vars)
			if content != "" {
				t.Logf("\t%s\tshould evaluate the expression with content %s", tests.Success, content)
			}

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
			contactEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedContactsEntityName)
			taskEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedTasksEntityName)

			contactItems, _ := item.List(tests.Context(), contactEntity.ID, db)
			taskItems, _ := item.List(tests.Context(), taskEntity.ID, db)

			vars, _ := node.MapToJSONB(map[string]string{contactEntity.ID: contactItems[0].ID})
			acts, _ := node.MapToJSONB(map[string]string{taskEntity.ID: taskItems[0].ID})

			node := node.Node{
				//Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				AccountID: schema.SeedAccountID,
				Variables: vars,
				Actuals:   acts,
				ActorID:   taskEntity.ID,
				Type:      node.Push,
			}
			eng := engine.Engine{
				Job: job.Job{},
			}
			_, err := eng.RunRuleEngine(tests.Context(), db, nil, node)
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
			contactEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedContactsEntityName)
			contactItems, _ := item.List(tests.Context(), contactEntity.ID, db)

			vars, _ := node.MapToJSONB(map[string]string{contactEntity.ID: contactItems[0].ID})
			acts, _ := node.MapToJSONB(map[string]string{contactEntity.ID: contactItems[1].ID}) //updatable-contact-id (Has blue-print of the values to be updated when triggered)

			node := node.Node{
				//Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				AccountID: schema.SeedAccountID,
				Variables: vars,
				Actuals:   acts,
				ActorID:   contactEntity.ID,
				Type:      node.Modify,
			}

			eng := engine.Engine{
				Job: job.Job{},
			}
			_, err := eng.RunRuleEngine(tests.Context(), db, nil, node)
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
		t.Log("\tWhen updating the event NPS score in contact")
		{
			contactEntity, _ := entity.RetrieveFixedEntity(tests.Context(), db, schema.SeedAccountID, schema.SeedContactsEntityName)
			contactItems, _ := item.List(tests.Context(), contactEntity.ID, db)
			i, _ := item.Retrieve(tests.Context(), contactEntity.ID, contactItems[0].ID, db)
			oldItemFields := i.Fields()
			newItemFields := i.Fields()
			newItemFields[schema.SeedFieldNPSKey] = 99
			item.UpdateFields(tests.Context(), db, contactEntity.ID, i.ID, newItemFields)
			// the above action will trigger this in the background thread
			flows, _ := flow.List(tests.Context(), []string{contactEntity.ID}, flow.FlowModeAll, db)
			dirtyFlows := flow.DirtyFlows(tests.Context(), flows, item.Diff(oldItemFields, newItemFields))

			eng := engine.Engine{
				Job: job.Job{},
			}
			errs := flow.Trigger(tests.Context(), db, nil, i.ID, dirtyFlows, eng)
			for _, err := range errs {
				if err != nil {
					t.Fatalf("\t%s should flow without error : %s.", tests.Failed, err)
				}
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
			n2 := "7e8e8aae-e13e-401d-8aa3-7200625bc7d2" //node-stage-2
			flowID := "ed58cf77-87e2-4d4c-a495-bfa7c808819f"
			f, err := flow.Retrieve(tests.Context(), flowID, db)
			contactItems, _ := item.List(tests.Context(), f.EntityID, db)
			eng := engine.Engine{
				Job: job.Job{},
			}
			err = flow.DirectTrigger(tests.Context(), db, nil, schema.SeedAccountID, flowID, n2, f.EntityID, contactItems[0].ID, eng)
			if err != nil {
				t.Fatalf("\t%s should flow with out error : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should flow with out error", tests.Success)

			afs, _ := flow.ActiveFlows(tests.Context(), []string{flowID}, db) //pipeline-id
			ans, _ := flow.ActiveNodes(tests.Context(), []string{flowID}, db) //pipeline-id
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

func TestGraphRuleRunner(t *testing.T) {
	db, teardown1 := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedPipelines(t, db)
	defer teardown1()
	residPool, teardown2 := tests.NewRedisUnit(t)
	defer teardown2()
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
			expression = fmt.Sprintf("{{%s.%s}} gt {50} && {{%s.%s}} eq {Bruce Wayne}", "380d2264-d4d7-444a-9f7a-ebca216bc67e", schema.SeedFieldNPSKey, "380d2264-d4d7-444a-9f7a-ebca216bc67e", schema.SeedFieldFNameKey)
			result := job.NewJabEngine().RunExpGrapher(tests.Context(), db, residPool, "3cf17266-3473-4006-984f-9325122678b7", expression)
			log.Println("result ------------------> ---> --> -> ", result)
			t.Logf("\t%s should pass with out fail", tests.Success)
		}
	}
}
