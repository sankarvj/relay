package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexANDOp emits a TokenANDOperation then returns
the lexer for begin.
*/
func LexANDOp(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.ANDOperation)
	lexer.Emit(lexertoken.TokenANDOperation)
	return LexBegin
}
