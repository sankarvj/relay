package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexAFSign emits a TokenAFSign then returns
the lexer for begin.
*/
func LexAFSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.AFSign)
	lexer.Emit(lexertoken.TokenAFSign)
	return LexBegin
}
