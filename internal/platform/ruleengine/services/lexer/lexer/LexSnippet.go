package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexSnippet emits a TokenSnippet with in left and right <>*/
func LexSnippet(lexer *Lexer) LexFn {
	for {
		if lexer.IsEOF() {
			return lexer.Errorf(errors.LexerErrorMissingSnippetClosure)
		}

		if strings.HasPrefix(lexer.InputToEnd(), lexertoken.RightSnippet) {
			lexer.Emit(lexertoken.TokenSnippet)
			return LexRightSnippet
		}

		lexer.Inc()
	}
}
