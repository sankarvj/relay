package engine_test

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestEmailRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
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
			rootNode, err := node.RootNode(branchNodeMap)
			log.Printf("The rootNode %v", rootNode)
			log.Println("The rootNode err", err)

			childNodes, err := node.ChildNodes(rootNode.ID, branchNodeMap)
			log.Printf("The childNodes %v", childNodes)
			log.Println("The childNodes err", err)

			// for i, n := range nodes {
			// 	log.Printf("node %d -- %v", i, n)
			// 	engine.RunRuleEngine(tests.Context(), db, n)
			// }
		}
	}
}
