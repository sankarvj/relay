package lexer

import (
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexANDOp emits a TokenANDOperation then returns
the lexer for begin.
*/
func LexANDOp(lexer *Lexer) LexFn {
	log.Println("Hello AND......")
	lexer.Pos += len(lexertoken.ANDOperation)
	lexer.Emit(lexertoken.TokenANDOperation)
	return LexBegin
}
