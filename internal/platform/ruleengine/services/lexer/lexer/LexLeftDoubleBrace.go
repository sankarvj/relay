package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexLeftDoubleBrace emits a TokenLeftDoubleBrace then returns
the lexer for a section.
*/
func LexLeftDoubleBrace(lexer *Lexer) LexFn {

	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingValuateOpener)
		}

		lexer.IgnoreWhitespace()

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftDoubleBraces) {
			lexer.Pos += len(lexertoken.LeftDoubleBraces)
			lexer.Emit(lexertoken.TokenLeftDoubleBrace)
			return LexValuate
		}
	}

}
