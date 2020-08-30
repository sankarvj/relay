package lexertoken

//EOF represents the end of file
const EOF rune = 0

//TOKENS represents the list of
const (
	LeftDoubleBraces   string = "{{"
	RightDoubleBraces  string = "}}"
	EqualSign          string = "eq"
	GTSign             string = "gt"
	LTSign             string = "lt"
	INSign             string = "in"
	LeftBrace          string = "{"
	RightBrace         string = "}"
	LeftSnippet        string = "<"
	RightSnippet       string = ">"
	LeftDoubleSnippet  string = "<<"
	RightDoubleSnippet string = ">>"
	ANDOperation       string = "&&"
	OROperation        string = "||"
)
