package lexer

import (
	"strings"
)

type lexType uint8

const (
	IDENTIFIER lexType = iota
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

func Lex(input string) []Item {
	var res []Item
	i := 0

	for i < len(input) {
		switch input[i] {
		case '+':
			fallthrough
		case '-':
			fallthrough
		case '*':
			fallthrough
		case '/':
			res = append(res, Item{Type: OPERATOR, Val: string(input[i])})
			break

		case '(':
			fallthrough
		case ')':
			res = append(res, Item{Type: SEPARATOR, Val: string(input[i])})
			break

		case '"':
			// String continues until next unescaped "
			var str string
			var escaped bool

			i++

			for i < len(input) {
				chr := input[i]

				if chr == '\\' {
					escaped = true
					i += 2
					continue
				}

				if !escaped && chr == '"' {
					break
				}

				if escaped {
					escaped = false
				}

				str += string(chr)
				i++
			}

			res = append(res, Item{Type: STRING, Val: str})
			break

		default:
			// NAME
			// Consists of a-z, parse until the last allowed char
			if input[i] >= 'a' && input[i] <= 'z' {
				name := ""

				for i < len(input) && input[i] >= 'a' && input[i] <= 'z' {
					name += string(input[i])
					i++
				}
				res = append(res, Item{Type: IDENTIFIER, Val: name})
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
			break
		}

		i++
	}

	res = append(res, Item{Type: EOF})

	return res
}
