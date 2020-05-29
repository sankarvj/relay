package lexer

import (
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexOROp emits a TokenOROperation then returns
the lexer for begin.
*/
func LexOROp(lexer *Lexer) LexFn {
	log.Println("Hello OR......")
	lexer.Pos += len(lexertoken.OROperation)
	lexer.Emit(lexertoken.TokenOROperation)
	return LexBegin
}
