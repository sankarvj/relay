package lexer

import (
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

/*
LexGibberish emits a TokenEOF after inc all the chars. But, why it does that?
ANS :: I wanna go through the usage trends and add some logic here in the near future
*/
func LexGibberish(lexer *Lexer) LexFn {
	for {
		if lexer.IsEOF() {
			lexer.Emit(lexertoken.TokenGibberish)
			lexer.Emit(lexertoken.TokenEOF)
			return nil
		}

		if lexer.IsWhitespace() {
			lexer.Emit(lexertoken.TokenGibberish)
			return LexBegin
		}

		lexer.Inc()
	}
}
