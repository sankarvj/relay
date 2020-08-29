package engine_test

import (
	"fmt"
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/item"
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
			e1 := schema.SeedEntityContactID
			e2 := schema.SeedEntityEmailID
			k1 := schema.SeedFieldKeyContactName
			i1 := schema.SeedItemContactID1
			i2 := schema.SeedItemEmailID

			vars, _ := node.MapToJSONB(map[string]string{e1: i1})
			acts, _ := node.MapToJSONB(map[string]string{e2: i2})

			node := node.Node{
				Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				Variables:  vars,
				Actuals:    acts,
				ActorID:    e2,
				Type:       node.Email,
			}

			_, err := engine.RunRuleEngine(tests.Context(), db, node)
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
			e1 := schema.SeedEntityContactID
			//k1 := schema.SeedFieldKeyContactName
			i1 := schema.SeedItemContactID1
			e2 := schema.SeedEntityTaskID
			i2 := schema.SeedItemTaskID2

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
			_, err := engine.RunRuleEngine(tests.Context(), db, node)
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
			e1 := schema.SeedEntityContactID
			//k1 := schema.SeedFieldKeyContactName
			i1 := schema.SeedItemContactID1
			i2 := schema.SeedItemContactUpdatableID

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
			_, err := engine.RunRuleEngine(tests.Context(), db, node)
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
			f1 := schema.SeedFlowID
			flow, _ := flow.Retrieve(tests.Context(), f1, db)
			log.Printf("The flow %v", flow)
			nodes, _ := node.List(tests.Context(), f1, db)

			branchNodeMap := node.BranceNodeMap(nodes)
			rootNodes := node.ChildNodes(node.Root, branchNodeMap)
			log.Printf("The rootNodes %v", rootNodes)

			childNodes := node.ChildNodes(rootNodes[0].ID, branchNodeMap)
			log.Printf("The childNodes %v", childNodes)

			// for i, n := range nodes {
			// 	log.Printf("node %d -- %v", i, n)
			// 	engine.RunRuleEngine(tests.Context(), db, n)
			// }
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
			e1 := schema.SeedEntityContactID
			i1 := schema.SeedItemContactID1
			i, _ := item.Retrieve(tests.Context(), i1, db)
			oldItemFields := i.Fields()
			newItemFields := i.Fields()
			newItemFields[schema.SeedFieldKeyContactMRR] = 99
			item.UpdateFields(tests.Context(), db, i1, newItemFields)
			flows, _ := flow.List(tests.Context(), e1, db)
			dirtyFlows := flow.DirtyFlows(tests.Context(), flows, oldItemFields, newItemFields)
			err := flow.Trigger(tests.Context(), db, i1, dirtyFlows)
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
			i1 := schema.SeedItemContactID1
			n2 := schema.SeedNodeID2
			err := flow.DirectTrigger(tests.Context(), db, n2, i1)
			if err != nil {
				t.Fatalf("\t%s should flow without error : %s.", tests.Failed, err)
			}
			t.Logf("\t%s should flow without error", tests.Success)

			afs, _ := flow.ActiveFlows(tests.Context(), []string{schema.SeedFlowID}, db)
			ans, _ := flow.ActiveNodes(tests.Context(), []string{schema.SeedFlowID}, db)
			log.Printf("afs >>>>>>>>>>>>>>>>>>>>>> %v", afs)
			log.Printf("ans >>>>>>>>>>>>>>>>>>>>>> %v", ans)

		}
	}
}
