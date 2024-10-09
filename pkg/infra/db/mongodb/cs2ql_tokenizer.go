package db

import (
	"fmt"
	"unicode"
)

// Token represents a single token in the CS2MatchQL query
type Token struct {
	Type  string // e.g., "KEYWORD", "IDENTIFIER", "OPERATOR", "LITERAL"
	Value string
}

// Tokenize parses a CS2MatchQL query string and returns a slice of tokens.
func Tokenize(query string) ([]Token, error) {
	var tokens []Token
	var currentToken Token

	for i := 0; i < len(query); {
		r := rune(query[i])

		switch {
		case unicode.IsSpace(r):
			// Skip whitespace
			i++
		case unicode.IsLetter(r) || r == '_':
			// Read identifier or keyword
			start := i
			for i < len(query) && (unicode.IsLetter(rune(query[i])) || unicode.IsDigit(rune(query[i])) || query[i] == '_') {
				i++
			}
			currentToken.Type = "IDENTIFIER"
			currentToken.Value = query[start:i]

			// Check for keywords
			switch currentToken.Value {
			case "match", "round", "event", "state", "stats":
				currentToken.Type = "KEYWORD"
			}

			tokens = append(tokens, currentToken)
			currentToken = Token{} // Reset for next token
		case unicode.IsDigit(r):
			// Read number literal
			start := i
			for i < len(query) && unicode.IsDigit(rune(query[i])) {
				i++
			}
			if i < len(query) && query[i] == '.' {
				i++ // Handle decimal point
				for i < len(query) && unicode.IsDigit(rune(query[i])) {
					i++
				}
			}
			currentToken.Type = "LITERAL_NUMBER"
			currentToken.Value = query[start:i]
			tokens = append(tokens, currentToken)
			currentToken = Token{}
		case r == '\'' || r == '"':
			// Read string literal
			quote := r
			i++
			start := i
			for i < len(query) && query[i] != byte(quote) {
				i++
			}
			if i == len(query) {
				return nil, fmt.Errorf("unterminated string literal")
			}
			currentToken.Type = "LITERAL_STRING"
			currentToken.Value = query[start:i]
			i++ // Skip the closing quote
			tokens = append(tokens, currentToken)
			currentToken = Token{}
		default:
			// Handle operators and other punctuation
			switch r {
			case '=', '!', '>', '<', '~', '(', ')', '[', ']', ',', '.', ':':
				currentToken.Type = "OPERATOR"
				currentToken.Value = string(r)
				tokens = append(tokens, currentToken)
				i++
				currentToken = Token{} // Reset for next token
			default:
				return nil, fmt.Errorf("unexpected character: %c", r)
			}
		}
	}

	return tokens, nil
}
