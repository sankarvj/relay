package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexLTSign emits a TokenLTSign then returns
the lexer for begin.
*/
func LexLTSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.LTSign)
	lexer.Emit(lexertoken.TokenLTSign)
	return LexBegin
}
