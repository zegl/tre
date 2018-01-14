package lexer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLexerSimpleAdd(t *testing.T) {
	r := Lex("aa + b")

	expected := []Item{
		{IDENTIFIER, "aa"},
		{OPERATOR, "+"},
		{IDENTIFIER, "b"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestLexerSimpleAddNumber(t *testing.T) {
	r := Lex("aa + 14")

	expected := []Item{
		{IDENTIFIER, "aa"},
		{OPERATOR, "+"},
		{NUMBER, "14"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestLexerSimpleCall(t *testing.T) {
	r := Lex("foo(bar)")

	expected := []Item{
		{IDENTIFIER, "foo"},
		{SEPARATOR, "("},
		{IDENTIFIER, "bar"},
		{SEPARATOR, ")"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestLexerSimpleCallWithString(t *testing.T) {
	r := Lex("foo(\"bar\")")

	expected := []Item{
		{IDENTIFIER, "foo"},
		{SEPARATOR, "("},
		{STRING, "bar"},
		{SEPARATOR, ")"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestString(t *testing.T) {
	r := Lex(`"bar"`)

	expected := []Item{
		{STRING, "bar"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestEscapedString(t *testing.T) {
	r := Lex(`"bar\""`)

	expected := []Item{
		{STRING, "bar\""},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestLexerSimpleCallWithTwoStrings(t *testing.T) {
	r := Lex(`foo("bar", "baz")`)

	expected := []Item{
		{IDENTIFIER, "foo"},
		{SEPARATOR, "("},
		{STRING, "bar"},
		{SEPARATOR, ","},
		{STRING, "baz"},
		{SEPARATOR, ")"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}

func TestLexerSimpleCallWithStringNum(t *testing.T) {
	r := Lex(`printf("%d\n", 123)`)

	expected := []Item{
		{IDENTIFIER, "printf"},
		{SEPARATOR, "("},
		{STRING, "%d\n"},
		{SEPARATOR, ","},
		{NUMBER, "123"},
		{SEPARATOR, ")"},
		{EOF, ""},
	}

	assert.Equal(t, expected, r)
}
