package errors

//Error constants for lexer errors
const (
	LexerErrorUnexpectedEOF         string = "Unexpected end of file"
	LexerErrorMissingValuateClosure string = "Missing a closing valuate braces"
	LexerErrorMissingValuateOpener  string = "Missing a opening valuate braces"
	LexerErrorMissingValueClosure   string = "Missing a closing value brace"
	LexerErrorMissingValueOpener    string = "Missing a opening value brace"
	LexerErrorMissingSnippetOpener  string = "Missing a opening snippet symbol"
	LexerErrorMissingSnippetClosure string = "Missing a closing snippet symbol"
)
