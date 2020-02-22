package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexRightBrace emits a TOKEN_RIGHT_BRACE then returns
the lexer for a begin.
*/
func LexRightBrace(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.RightBrace)
	lexer.Emit(lexertoken.TokenRightBrace)
	return LexBegin
}
