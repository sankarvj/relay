package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexLeftSnippet emits a TokenLeftSnippet then returns
the lexer for a section.
*/
func LexLeftSnippet(lexer *Lexer) LexFn {

	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingSnippetOpener)
		}

		lexer.IgnoreWhitespace()

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftSnippet) {
			lexer.Pos += len(lexertoken.LeftSnippet)
			lexer.Emit(lexertoken.TokenLeftSnippet)
			return LexSnippet
		}
	}

}
