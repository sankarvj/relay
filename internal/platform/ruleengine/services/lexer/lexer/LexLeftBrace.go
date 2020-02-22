package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexLeftBrace emits a TOKEN_LEFT_BRACE then returns
the lexer for a section.
*/
func LexLeftBrace(lexer *Lexer) LexFn {

	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingValueOpener)
		}

		lexer.IgnoreWhitespace()

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftBrace) {
			lexer.Pos += len(lexertoken.LeftBrace)
			lexer.Emit(lexertoken.TokenLeftBrace)
			return LexValue
		}

	}

}
