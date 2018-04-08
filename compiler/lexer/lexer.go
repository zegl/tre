package lexer

import (
	"fmt"
	"strings"
)

type lexType uint8

const (
	IDENTIFIER lexType = iota
	KEYWORD
	NUMBER
	STRING
	OPERATOR
	SEPARATOR

	EOF
	EOL
)

type Item struct {
	Type lexType
	Val  string
	Line int
}

func (i Item) String() string {
	var t string
	switch i.Type {
	case IDENTIFIER:
		t = "IDENTIFIER"
	case KEYWORD:
		t = "KEYWORD"
	case NUMBER:
		t = "NUMBER"
	case STRING:
		t = "STRING"
	case OPERATOR:
		t = "OPERATOR"
	case SEPARATOR:
		t = "SEPARATOR"
	case EOF:
		t = "EOF"
	case EOL:
		t = "EOL"
	}

	return fmt.Sprintf("{Type:%s, Val:%s, Line:%d}", t, i.Val, i.Line)
}

var operations = map[string]struct{}{
	"+":  {},
	"-":  {},
	"*":  {}, // Multiplication and dereferencing
	"/":  {},
	"<":  {},
	">":  {},
	"<=": {},
	">=": {},
	"==": {},
	"!=": {},
	"=":  {},
	"!":  {},
	":":  {}, // is not really a operation. Is only defined here so that := can be found.
	":=": {},
	".":  {},
	"[":  {},
	"]":  {},
	"&":  {},
}

var separators = map[string]struct{}{
	")": {},
	"(": {},
	"}": {},
	"{": {},
	",": {},
	";": {},
}

func Lex(inputFullSource string) []Item {
	var res []Item

	for line, input := range strings.Split(inputFullSource, "\n") {
		if line > 0 {
			res = append(res, Item{Type: EOL, Line: line})
		}

		// Lines starts at 1
		line = line + 1

		i := 0

		for i < len(input) {

			// Comment, until end of line or end of file
			if input[i] == '/' && input[i+1] == '/' {
				break
			}

			if _, ok := operations[string(input[i])]; ok {

				// Operators can be 1 or 2 characters
				if len(input) > i+1 {
					if _, ok := operations[string(input[i])+string(input[i+1])]; ok {
						res = append(res, Item{Type: OPERATOR, Val: string(input[i]) + string(input[i+1]), Line: line})
						i++
						i++
						continue
					}
				}

				res = append(res, Item{Type: OPERATOR, Val: string(input[i]), Line: line})
				i++
				continue
			}

			if _, ok := separators[string(input[i])]; ok {
				res = append(res, Item{Type: SEPARATOR, Val: string(input[i]), Line: line})
				i++
				continue
			}

			if input[i] == '"' {
				// String continues until next unescaped "
				var str string

				i++

				for i < len(input) {
					if input[i] == '"' {
						break
					}

					// parse escape sequences
					if input[i] == '\\' {
						if esc, ok := escapeSequences[string(input[i])+string(input[i+1])]; ok {
							str += esc
							i += 2
							continue
						}
					}

					str += string(input[i])
					i++
				}

				i++
				res = append(res, Item{Type: STRING, Val: str, Line: line})
				continue
			}

			// NAME
			// Consists of a-z, parse until the last allowed char
			if (input[i] >= 'a' && input[i] <= 'z') || (input[i] >= 'A' && input[i] <= 'Z') {
				name := ""

				for i < len(input) && ((input[i] >= 'a' && input[i] <= 'z') ||
					(input[i] >= 'A' && input[i] <= 'Z') ||
					(input[i] >= '0' && input[i] <= '9') ||
					input[i] == '_') {
					name += string(input[i])
					i++
				}

				if _, ok := keywords[name]; ok {
					res = append(res, Item{Type: KEYWORD, Val: name, Line: line})
				} else {
					res = append(res, Item{Type: IDENTIFIER, Val: name, Line: line})
				}

				continue
			}

			// NUMBER
			// 0-9
			if input[i] >= '0' && input[i] <= '9' {
				val := ""
				for i < len(input) && input[i] >= '0' && input[i] <= '9' {
					val += string(input[i])
					i++
				}
				res = append(res, Item{Type: NUMBER, Val: val, Line: line})
				continue
			}

			// Whitespace (ignore)
			if len(strings.TrimSpace(string(input[i]))) == 0 {
				i++
				continue
			}

			panic("Unexpected char in Lexer: " + string(input[i]))
		}
	}

	res = append(res, Item{Type: EOF})

	return res
}
