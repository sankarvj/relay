package rule_test

import (
	"fmt"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/rule"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestRuleRunner(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	tests.SeedData(t, db)
	defer teardown()
	t.Log("Given the need to run a rule for the sample expression.")
	{
		t.Log("\tWhen running a rule")
		{
			e1 := schema.SeedEntityContactID
			e2 := schema.SeedEntityEmailID
			k1 := schema.SeedFieldKeyContactName
			i1 := schema.SeedItemContactID1
			i2 := schema.SeedItemEmailID
			sampleRule := fmt.Sprintf("{{%s.%s}} eq {Vijay}  <%s.%s>", e1, k1, e2, i2)
			rule.RunRuleEngine(tests.Context(), db, sampleRule, map[string]string{e1: i1})
		}
	}

}