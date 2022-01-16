package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexGTSign emits a TokenGTSign then returns
the lexer for begin.
*/
func LexLKSign(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.LikeSign)
	lexer.Emit(lexertoken.TokenLKSign)
	return LexBegin
}
