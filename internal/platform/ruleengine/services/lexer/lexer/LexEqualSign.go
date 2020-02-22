package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexEqualSign emits a TokenEqualSign then returns
the lexer for begin.
*/
func LexEqualSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.EqualSign)
	lexer.Emit(lexertoken.TokenEqualSign)
	return LexBegin
}
