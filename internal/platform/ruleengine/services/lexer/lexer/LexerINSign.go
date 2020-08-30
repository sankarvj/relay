package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexINSign emits a TokenINSign then returns
the lexer for begin.
*/
func LexINSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.INSign)
	lexer.Emit(lexertoken.TokenINSign)
	return LexBegin
}
