package lexertoken

//TokenType is the special type of token
type TokenType int

//Tokens for different types
const (
	TokenError TokenType = iota
	TokenEOF

	TokenLeftDoubleBrace
	TokenRightDoubleBrace
	TokenLeftBrace
	TokenRightBrace
	TokenLeftSnippet
	TokenRightSnippet

	TokenEqualSign
	TokenANDOperation
	TokenOROperation
	TokenValuate
	TokenValue
	TokenGibberish
	TokenSnippet
)
