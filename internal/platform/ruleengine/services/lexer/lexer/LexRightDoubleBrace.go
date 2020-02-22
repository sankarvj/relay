package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexRightDoubleBrace emits a TokenRightDoubleBrace then returns
the lexer for a begin.
*/
func LexRightDoubleBrace(lexer *Lexer) LexFn {
	lexer.Pos += len(lexertoken.RightDoubleBraces)
	lexer.Emit(lexertoken.TokenRightDoubleBrace)
	return LexBegin
}
