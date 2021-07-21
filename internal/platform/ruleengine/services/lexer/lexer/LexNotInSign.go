package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexNotINSign emits a TokenNotINSign then returns
the lexer for begin.
*/
func LexNotINSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.NotINSign)
	lexer.Emit(lexertoken.TokenNotINSign)
	return LexBegin
}
