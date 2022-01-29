package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexBWSign emits a TokenBWSign then returns
the lexer for begin.
*/
func LexBWSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.BWSign)
	lexer.Emit(lexertoken.TokenBWSign)
	return LexBegin
}
