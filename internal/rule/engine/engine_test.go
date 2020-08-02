package engine_test

import (
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
		t.Log("\tWhen running a send email engine")
		{
			e1 := schema.SeedEntityContactID
			e2 := schema.SeedEntityEmailID
			//k1 := schema.SeedFieldKeyContactName
			i1 := schema.SeedItemContactID1
			i2 := schema.SeedItemEmailID

			vars, _ := node.MapToJSONB(map[string]string{e1: i1})
			acts, _ := node.MapToJSONB(map[string]string{e2: i2})

			node := node.Node{
				//Expression: fmt.Sprintf("{{%s.%s}} eq {Vijay}", e1, k1),
				Variables: vars,
				Actuals:   acts,
				ActorID:   e2,
				Type:      node.Email,
			}

			engine.RunRuleEngine(tests.Context(), db, node)
		}
	}
}

func TestCreateRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedWorkFlows(t, db)
	defer teardown()
	t.Log("Given the need to run the engine to create new item")
	{
		t.Log("\tWhen running a create item engine")
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
			engine.RunRuleEngine(tests.Context(), db, node)
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
		t.Log("\tWhen running update item engine")
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

			engine.RunRuleEngine(tests.Context(), db, node)
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
		t.Log("\tWhen running a flow 'The flow'")
		{
			f1 := schema.SeedFlowID
			flow, _ := flow.Retrieve(tests.Context(), f1, db)
			log.Printf("The flow %v", flow)
			nodes, _ := node.List(tests.Context(), f1, db)

			branchNodeMap := node.BranceNodeMap(nodes)
			rootNodes := node.ChildNodes("root", branchNodeMap)
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
			//log.Println("oldItemFields", oldItemFields)
			//log.Println("newItemFields", newItemFields)
			flows, _ := flow.List(tests.Context(), e1, db)
			dirtyFlows := flow.DirtyFlows(tests.Context(), flows, oldItemFields, newItemFields)
			//log.Printf("The lazyFlows %v", lazyFlows)
			flow.Trigger(tests.Context(), i1, dirtyFlows, db)

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
		t.Log("\tWhen updating the event mrr in contact1")
		{
			e1 := schema.SeedEntityContactID
			i1 := schema.SeedItemContactID1
			n1 := schema.SeedNodeID2
			n, _ := node.Retrieve(tests.Context(), n1, db)
			flow.DirectTrigger(tests.Context(), db, *n, e1, i1, 3)

			afs, _ := flow.ActiveFlows(tests.Context(), []string{n.FlowID}, db)
			ans, _ := flow.ActiveNodes(tests.Context(), []string{n.FlowID}, db)
			log.Printf("afs >>>>>>>>>>>>>>>>>>>>>> %v", afs)
			log.Printf("ans >>>>>>>>>>>>>>>>>>>>>> %v", ans)

		}
	}
}
