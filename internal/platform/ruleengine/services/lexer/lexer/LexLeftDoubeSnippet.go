package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexLeftDoubleSnippet emits a TokenLeftDoubleBrace then returns
the lexer for a section.
*/
func LexLeftDoubleSnippet(lexer *Lexer) LexFn {

	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingQueryOpener)
		}

		lexer.IgnoreWhitespace()

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftDoubleSnippet) {
			lexer.Pos += len(lexertoken.LeftDoubleSnippet)
			lexer.Emit(lexertoken.TokenLeftDoubleSnippet)
			return LexQuery
		}
	}

}
