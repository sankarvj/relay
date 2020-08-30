# Rule Engine

The rule engine internally has a lexer and parser. Lexer find the tokens from the expression, the parser works on that token to evaluate the actual value. The engine runs on top of it. 

For the engine to run you need to pass three things, a expression to be evalated and the variables map and the actuals map. 

## Variables Map - Worker usecase
Executing the rule is dynamic, so the expression will not contain the itemID by default. If {{e1.f1}} is the expression then the e1 will be replaced by the item provided in the variables map. 

Inorder to reduce the DB calls we store the itemMap itself instead of ID in the variables map if once processed...... to be continue......

## Engine
The engine is the outermost wrapper which handles three important tasks, which are:-
1. Aggeregating AND/OR condition results
2. Call the operators to evaluate the expression (more details below)
3. Send back the evaluted expression
4. Send back the pos/neg trigger

## Lexer
The lexer is the token fetcher. It fetches the token and given that to lexer.
Following are the valid lexter tokens:-
`{{e1.f1}}` - evaluatable operands (usually in entity.field format)
`{1000}` - non-evaluatable operands (usually a hardcoded value of string/number)
`<age=50>` - an actionable snippet. Yet to explore.
`&&,||` - conditions
`eq,lt,gt,in` - operators which act on left and right operands

## Parser
The parser is the one which converts the lexer tokens into values with the help of worker and operators, for example
lets say we passed the input map as:- 
{
    e1: {
        f1: 1000
    }
}
`{{e1.f1}}` will be converted to 1000 
`{1000}` will be converted to 1000

## How the engine call the specific operators to evaluate the specific expression?
for example, inorder to evaluate `{{e1.f1}} eq 1000` the engine has to call the compare operator with (1000,1000) as the two operands.

## Steps to add new token
1. Add the new token in lexertoken/Tokens.go
2. Add the new token type enum lexertoken/TokenType.go
3. Add a new func in the lexer 
4. Add the new case in startExecutingLexer/startParsingLexer whichever in applicable
5. Add the new Lex func to the Lex Begin
6. That's all folks, your lexer is ready to find the new token 


