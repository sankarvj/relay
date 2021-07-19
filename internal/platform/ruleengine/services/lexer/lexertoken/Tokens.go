package lexertoken

//EOF represents the end of file
const EOF rune = 0

//TOKENS represents the list of
const (
	LeftDoubleBraces  string = "{{"
	RightDoubleBraces string = "}}"
	LeftBrace         string = "{"
	RightBrace        string = "}"
	EqualSign         string = "eq"
	GTSign            string = "gt"
	LTSign            string = "lt"
	BFSign            string = "bf"
	AFSign            string = "af"
	INSign            string = "in"

	LeftSnippet        string = "<"
	RightSnippet       string = ">"
	LeftDoubleSnippet  string = "<<"
	RightDoubleSnippet string = ">>"
	ANDOperation       string = "&&"
	OROperation        string = "||"
)
