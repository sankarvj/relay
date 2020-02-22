package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexRightSnippet emits a TOKEN_RIGHT_BRACKET then returns
the lexer for a begin.
*/
func LexRightSnippet(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.RightSnippet)
	lexer.Emit(lexertoken.TokenRightSnippet)
	return LexBegin
}
