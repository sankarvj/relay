package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexRightDoubleSnippet emits a TokenRightDoubleSnippet then returns
the lexer for a begin.
*/
func LexRightDoubleSnippet(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.RightDoubleSnippet)
	lexer.Emit(lexertoken.TokenRightDoubleSnippet)
	return LexBegin
}
