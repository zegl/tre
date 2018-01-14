package lexer

var escapeSequences = map[string]string{
	`\"`: "\"",
	`\n`: "\n",
}
