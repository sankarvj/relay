package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexGTSign emits a TokenGTSign then returns
the lexer for begin.
*/
func LexGTSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.GTSign)
	lexer.Emit(lexertoken.TokenGTSign)
	return LexBegin
}
