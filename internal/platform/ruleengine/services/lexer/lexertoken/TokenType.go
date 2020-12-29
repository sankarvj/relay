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
	TokenLeftDoubleSnippet
	TokenRightDoubleSnippet
	TokenSnippet

	TokenEqualSign
	TokenGTSign
	TokenLTSign
	TokenINSign
	TokenBFSign
	TokenAFSign

	TokenANDOperation
	TokenOROperation

	TokenValuate
	TokenQuery
	TokenValue
	TokenGibberish
)
