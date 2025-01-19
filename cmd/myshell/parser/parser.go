package parser

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Parser struct {
	input            []rune
	position         int
	isInSingleQuotes bool
	isInDoubleQuotes bool
}

var SpecialCharacters = []rune{'\\', '"', '$', '`'}

func New(input string) Parser {
	return Parser{
		input:            []rune(input),
		position:         0,
		isInSingleQuotes: false,
		isInDoubleQuotes: false,
	}
}

func (p *Parser) Print() {
	fmt.Printf("Parser{input: %s, position: %d, current: [%c]}\n", string(p.input), p.position, p.Current())
}

func (p *Parser) Parse() []string {

	var tokens []string
	var parts []string

	for p.HasCurrent() {
		var s string = ""

		// p.Print()

		// Check if a space which separates the tokens. If not a space
		// then parse correct and append to the current token.
		if unicode.IsSpace(p.Current()) {
			if parts != nil {
				tokens = append(tokens, strings.Join(parts, ""))
				parts = nil
			}
			p.Next()
		} else {
			switch p.Current() {
			case '\'':
				s = p.ParseSingleQuotes()
			case '"':
				s = p.ParseDoubleQuotes()
			default:
				s = p.ParseWord()
			}
		}

		if s != "" {
			parts = append(parts, s)
		}
	}

	if parts != nil {
		tokens = append(tokens, strings.Join(parts, ""))
	}
	return tokens
}

func (p *Parser) Backtrack() {
	if p.position > 0 {
		p.position--
	}
}

func (p *Parser) Next() {
	p.position++
}

func (p *Parser) HasCurrent() bool {
	return p.position < len(p.input)
}

func (p *Parser) Current() rune {
	if p.HasCurrent() {
		return p.input[p.position]
	}

	return 0
}

func (p *Parser) ParseWord() string {
	sb := strings.Builder{}

	for p.HasCurrent() && !slices.Contains([]rune{'\'', '"', ' '}, p.Current()) {
		if p.Current() == '\\' {
			p.Next()
		}

		sb.WriteRune(p.Current())
		p.Next()
	}

	return sb.String()
}

func (p *Parser) ParseSingleQuotes() string {
	// https://www.gnu.org/software/bash/manual/bash.html#Single-Quotes
	p.Next() // Consume the opening single quote

	sb := strings.Builder{}
	for p.HasCurrent() {
		if p.Current() == '\'' {

			// Check if two single quotes are next to each other
			// If so, the consume both effectively concatting the two strings
			p.Next()
			if p.Current() == '\'' {
				p.Next()
				continue

			}

			// Give back the character consumed since the second one wasn't
			// another single quote
			p.Backtrack()
			break
		}

		sb.WriteRune(p.Current())
		p.Next()
	}

	if p.Current() != '\'' {
		panic("unmatched single quote")
	}
	p.Next() // Consume the closing single quote

	return sb.String()
}

func (p *Parser) ParseDoubleQuotes() string {
	// https://www.gnu.org/software/bash/manual/bash.html#Double-Quotes
	p.Next() // Consume the opening double quote

	sb := strings.Builder{}
	for p.HasCurrent() {
		if p.Current() == '"' {

			// Check if two double quotes are next to each other
			// If so, the consume both effectively concatenating the two strings
			p.Next()
			if p.Current() == '"' {
				p.Next()
				continue
			}

			// Give back the character consumed since the second one wasn't
			// another double quote
			p.Backtrack()
			break
		}

		if p.Current() == '\\' {
			// https://www.gnu.org/software/bash/manual/bash.html#Escape-Character
			// If the character after the backslash is a special character, then
			// consume the backslash and only append the special character
			p.Next()
			if !slices.Contains(SpecialCharacters, p.Current()) {
				// It wasn't a special character so go one back and append the backslash as is
				p.Backtrack()
			}
		}

		sb.WriteRune(p.Current())
		p.Next()
	}

	if p.Current() != '"' {
		panic("unmatched double quote")
	}
	p.Next() // Consume the closing double quote

	return sb.String()
}
