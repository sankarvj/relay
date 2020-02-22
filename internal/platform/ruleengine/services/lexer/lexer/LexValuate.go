package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexValuate emits a TokenValuate with in left and right double braces*/
func LexValuate(lexer *Lexer) LexFn {
	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingValuateClosure)
		}

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.RightDoubleBraces) {
			lexer.Emit(lexertoken.TokenValuate)
			return LexRightDoubleBrace
		}

		lexer.Inc()
	}
}
