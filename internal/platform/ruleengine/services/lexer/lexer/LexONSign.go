package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexONSign emits a TokenONSign then returns
the lexer for begin.
*/
func LexONSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.ONSign)
	lexer.Emit(lexertoken.TokenONSign)
	return LexBegin
}
