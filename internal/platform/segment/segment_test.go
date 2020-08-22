package segment_test

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/segment"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

var (
	conditions = []segment.Condition{
		segment.Condition{
			Operator: ">",
			Key:      "age",
			Type:     "N",
			Value:    "40",
		},
		segment.Condition{
			Operator: "=",
			Key:      "name",
			Type:     "S",
			Value:    "Siva",
		},
	}
	seg = segment.Segment{
		Match:      segment.MatchAll,
		Conditions: conditions,
	}
)

func TestParseSegmentForGraph(t *testing.T) {
	t.Log(" Given the need to parse the segment into graph query")
	{
		t.Log("\twhen parsing AND conditions")
		{
			q, err := segment.ParseSegmentForGraph("xyz", seg)

			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
			log.Println("q ------> ", q)
		}
	}
}
