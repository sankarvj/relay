package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexValue handles the actual eval value
*/
func LexValue(lexer *Lexer) LexFn {
	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingValueClosure)
		}

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.RightBrace) {
			lexer.Emit(lexertoken.TokenValue)
			return LexRightBrace
		}

		lexer.Inc()
	}
}
