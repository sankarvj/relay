# Rule Engine

The rule engine internally has a lexer and parser. Lexer find the tokens from the expression, the parser works on that token to evaluate the actual value. The engine runs on top of it. 

## Engine
The engine is the outermost wrapper which handles three important tasks, which are:-
1. Aggeregating AND/OR condition results
2. Send back the evaluted expression
3. Send back the pos/neg trigger

## Lexer
The lexer is the token fetcher. Definition of tokens
`{{e1.f1.val}}` - evaluatable operands (usually in entity.field.value format)
`{vijay}` - non-evaluatable operands (usually a hardcoded value of string/number)
`<age=50>` - an actionable snippet. Yet to explore.
`&&,||` - conditions
`eq,lt,gt` - operators which act on left and right operands

## Parser
The rule engine expects the 

