package entity

const (
	MatchAll string = "ALL"
	MatchAny        = "ANY"
)

type Segment struct {
	Match           string
	FieldConditions []Field
}
