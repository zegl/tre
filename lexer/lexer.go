package lexer

import "strings"

type lexType uint8

const (
	IDENTIFIER lexType = iota
	KEYWORD
	NUMBER
	STRING
	OPERATOR
	SEPARATOR
	EOF
)

type Item struct {
	Type lexType
	Val  string
}

var operations = map[string]struct{}{
	"+":  {},
	"-":  {},
	"*":  {},
	"/":  {},
	"<":  {},
	">":  {},
	"<=": {},
	">=": {},
	"=":  {},
	"==": {},
	"!=": {},
}

var separators = map[string]struct{}{
	")": {},
	"(": {},
	"}": {},
	"{": {},
	",": {},
}

func Lex(input string) []Item {
	var res []Item
	i := 0

	for i < len(input) {

		if _, ok := operations[string(input[i])]; ok {
			// Operators can be 1 or 2 characters
			if _, ok := operations[string(input[i])+string(input[i+1])]; ok {
				res = append(res, Item{Type: OPERATOR, Val: string(input[i]) + string(input[i+1])})
				i++
			} else {
				res = append(res, Item{Type: OPERATOR, Val: string(input[i])})
			}

			i++
			continue
		}

		if _, ok := separators[string(input[i])]; ok {
			res = append(res, Item{Type: SEPARATOR, Val: string(input[i])})
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
			res = append(res, Item{Type: STRING, Val: str})
			continue
		}

		// NAME
		// Consists of a-z, parse until the last allowed char
		if input[i] >= 'a' && input[i] <= 'z' {
			name := ""

			for i < len(input) && input[i] >= 'a' && input[i] <= 'z' {
				name += string(input[i])
				i++
			}

			if _, ok := keywords[name]; ok {
				res = append(res, Item{Type: KEYWORD, Val: name})
			} else {
				res = append(res, Item{Type: IDENTIFIER, Val: name})
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
			res = append(res, Item{Type: NUMBER, Val: val})
			continue
		}

		// Whitespace (ignore)
		if len(strings.TrimSpace(string(input[i]))) == 0 {
			i++
			continue
		}

		panic("Unexpected char in Lexer: " + string(input[i]))
	}

	res = append(res, Item{Type: EOF})

	return res
}
