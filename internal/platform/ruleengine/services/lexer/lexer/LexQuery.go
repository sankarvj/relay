package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexQuery emits a TokenQuery with in left and right double snippet*/
func LexQuery(lexer *Lexer) LexFn {
	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingQueryClosure)
		}

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.RightDoubleSnippet) {
			lexer.Emit(lexertoken.TokenQuery)
			return LexRightDoubleSnippet
		}

		lexer.Inc()
	}
}
