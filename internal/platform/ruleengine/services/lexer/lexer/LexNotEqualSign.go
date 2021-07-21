package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexNotEqualSign emits a TokenNotEqualSign then returns
the lexer for begin.
*/
func LexNotEqualSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.NotEqualSign)
	lexer.Emit(lexertoken.TokenNotEqualSign)
	return LexBegin
}
