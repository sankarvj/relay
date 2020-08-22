package segment

import (
	"fmt"
	"strings"
)

//Const to describe the match format
const (
	MatchAll string = "ALL"
	MatchAny        = "ANY"
)

type Segment struct {
	Match      string
	Conditions []Condition
}

type Condition struct {
	Operator string
	Key      string
	Value    string
	Type     string
	Segment  *Segment // as of now no use... use it for complex conditions. Mostly try to solve in the lexer and remove it here
}

func ParseSegmentForGraph(label string, seg Segment) (string, error) {
	if seg.Match == MatchAny { //OR

	}
	var s []string
	var c string
	for _, condition := range seg.Conditions {
		if condition.Type == "S" {
			c = fmt.Sprintf("%s.%s %s `%s`", label, condition.Key, condition.Operator, condition.Value)
		} else if condition.Type == "N" {
			c = fmt.Sprintf("%s.%s %s %s", label, condition.Key, condition.Operator, condition.Value)
		} else {
			c = fmt.Sprintf("%s.%s %s %s", label, condition.Key, condition.Operator, condition.Value)
		}
		s = append(s, c)
	}
	return strings.Join(s, ", "), nil

}
