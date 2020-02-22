package lexertoken

import (
	"fmt"
)

//Token describes type and the value while lexing each section
type Token struct {
	Type  TokenType
	Value string
}

//String makes token type with string capabilities
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"

	case TokenError:
		return t.Value
	}

	return fmt.Sprintf("%q", t.Value)
}
