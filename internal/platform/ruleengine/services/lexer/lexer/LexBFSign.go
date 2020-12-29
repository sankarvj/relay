package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexBFSign emits a TokenBFSign then returns
the lexer for begin.
*/
func LexBFSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.BFSign)
	lexer.Emit(lexertoken.TokenBFSign)
	return LexBegin
}
