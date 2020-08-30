package ruler

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestOperators(t *testing.T) {
	//signalsChan wait to receive work and action triggers until the run completes
	t.Log(" Given the need test the operators(GT/LT/EQ) in the expression")
	{
		t.Log("\twhen evaluating lt")
		{
			var triggered bool
			exp := `{{e1.appinfo.index}} lt {{e2.index}} && {{e1.appinfo.index}} lt {{e2.index}} <e3.status=e2.version>`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, true, signalsChan)
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInput(work.Expression)
				case PosExecutor:
					triggered = true
					t.Logf("\t%s should receive positive trigger after evaluting lt", tests.Success)
				}
			}
			if !triggered {
				t.Fatalf("\t%s should receive positive trigger after evaluting lt", tests.Failed)
			}
		}

		t.Log("\twhen evaluating eq")
		{
			var triggered bool
			exp := `{{e1.appinfo.version}} eq {{e2.version}} && {{e1.appinfo.version}} eq {{e2.version}} <e3.status=e2.version>`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, true, signalsChan)
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInput(work.Expression)
				case PosExecutor:
					triggered = true
					t.Logf("\t%s should receive positive trigger after evaluting eq", tests.Success)
				}
			}
			if !triggered {
				t.Fatalf("\t%s should receive positive trigger after evaluting eq", tests.Failed)
			}
		}
	}
}

func TestContentParser(t *testing.T) {
	t.Log(" Given the need evaluate/parse the given expression")
	{
		t.Log("\twhen evaluating subject line : ")
		{
			var content string
			exp := `Hello matty {{e1.appinfo.version}}. How are you?`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, false, signalsChan)
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInput(work.Expression)
				case Content:
					content = work.Expression
				}
			}
			if content == "Hello matty 2 . How are you?" {
				t.Logf("\t%s should parse the expression with proper value", tests.Success)
			} else {
				t.Fatalf("\t%s should parse the expression with proper value. Parsed content: %s", tests.Failed, content)
			}
		}
	}

}

func TestExpressionWithListOperands(t *testing.T) {
	t.Log("Given the need run the rule engine for the expression consists list")
	{
		t.Log("\twhen the right operand exists in the left list : ")
		{
			exp := `{{e1.supports}} in {sdk2}`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, true, signalsChan)
			triggered := false
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInputWithList(work.Expression)
				case PosExecutor:
					triggered = true
					t.Logf("\t%s should receive positive trigger", tests.Success)
				}
			}
			if !triggered {
				t.Fatalf("\t%s should receive positive trigger", tests.Failed)
			}
		}

		t.Log("\twhen the right operand does not exists in the left list")
		{
			exp := `{{e1.supports}} in {sdk3}`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, true, signalsChan)
			triggered := false
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInputWithList(work.Expression)
				case PosExecutor:
					triggered = true
					t.Fatalf("\t%s should not revice positive trigger", tests.Failed)
				}
			}
			if !triggered {
				t.Logf("\t%s should not receive positive trigger", tests.Success)
			}
		}
	}

}

func TestQuerySnippet(t *testing.T) {
	//signalsChan wait to receive work and action triggers until the run completes
	t.Log(" Given the need test the query snippets in the expression")
	{
		t.Log("\twhen evaluating query")
		{
			var triggered bool
			exp := `{{e1.appinfo.index}} lt {{e2.index}} && <<helloquery>>`
			t.Log("expression --> ", exp)
			signalsChan := make(chan Work)
			go Run(exp, true, signalsChan)
			for work := range signalsChan {
				switch work.Type {
				case Worker:
					work.Resp <- workerMockInput(work.Expression)
				case PosExecutor:
					triggered = true
					t.Logf("\t%s should receive positive trigger after evaluting lt", tests.Success)
				}
			}
			if !triggered {
				t.Fatalf("\t%s should receive positive trigger after evaluting lt", tests.Failed)
			}
		}
	}
}

//mock inputs for workers
func workerMockInput(exp string) map[string]interface{} {
	key := FetchEntityID(exp)
	if key == "e1" {
		return map[string]interface{}{
			"e1": map[string]interface{}{
				"artifact": 1,
				"appinfo": map[string]interface{}{
					"version": 2,
					"index":   99,
				},
			},
		}
	} else {
		return map[string]interface{}{
			"e2": map[string]interface{}{
				"version": 2,
				"index":   100,
			},
		}
	}
}

func workerMockInputWithList(exp string) map[string]interface{} {
	key := FetchEntityID(exp)
	if key == "e1" {
		return map[string]interface{}{
			"e1": map[string]interface{}{
				"artifact": 1,
				"supports": []interface{}{"sdk1", "sdk2"},
				"appinfo": map[string]interface{}{
					"version": 2,
					"index":   99,
				},
			},
		}
	} else {
		return map[string]interface{}{
			"e2": map[string]interface{}{
				"version": 2,
				"index":   100,
			},
		}
	}
}
