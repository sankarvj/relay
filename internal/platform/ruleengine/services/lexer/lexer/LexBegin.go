package lexer

import (
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexBegin starts everything off. It determines if we are
beginning with a key/value assignment or a section.
*/
func LexBegin(lexer *Lexer) LexFn {
	lexer.IgnoreWhitespace()

	if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftDoubleBraces) {
		return LexLeftDoubleBrace
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftBrace) {
		return LexLeftBrace
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftDoubleSnippet) {
		return LexLeftDoubleSnippet
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.EqualSign) {
		return LexEqualSign
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.INSign) {
		return LexINSign
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.GTSign) {
		return LexGTSign
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LTSign) {
		return LexLTSign
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.ANDOperation) {
		return LexANDOp
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.OROperation) {
		return LexOROp
	} else if strings.HasPrefix(lexer.InputToEnd(), lexertoken.LeftSnippet) {
		return LexLeftSnippet
	} else {
		return LexGibberish
	}
}
