package parser

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Parser struct {
	Input            []rune
	Position         int
	IsInSingleQuotes bool
	IsInDoubleQuotes bool
}

var SpecialCharacters = []rune{'\\', '"', '$', '`'}

func New(input string) Parser {
	return Parser{
		Input:            []rune(input),
		Position:         0,
		IsInSingleQuotes: false,
		IsInDoubleQuotes: false,
	}
}

func (p *Parser) Print() {
	fmt.Printf("Parser{input: %s, position: %d, current: [%c]}\n", string(p.Input), p.Position, p.Current())
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
				p.IsInSingleQuotes = true
				s = p.ParseQuotes()
			case '"':
				p.IsInDoubleQuotes = true
				s = p.ParseQuotes()
			default:
				s = p.ParseWord()
			}
		}

		p.IsInDoubleQuotes = false
		p.IsInSingleQuotes = false

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
	if p.Position > 0 {
		p.Position--
	}
}

func (p *Parser) Next() {
	p.Position++
}

func (p *Parser) HasCurrent() bool {
	return p.Position < len(p.Input)
}

func (p *Parser) Current() rune {
	if p.HasCurrent() {
		return p.Input[p.Position]
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

func (p *Parser) ParseQuotes() string {
	// https://www.gnu.org/software/bash/manual/bash.html#Single-Quotes
	// https://www.gnu.org/software/bash/manual/bash.html#Double-Quotes
	p.Next() // Consume the opening quote

	mc := '\''
	if p.IsInDoubleQuotes {
		mc = '"'
	}

	sb := strings.Builder{}
	for p.HasCurrent() {
		if p.Current() == mc {

			// Check if two quotes are next to each other
			// If so, the consume both effectively concatenating the two strings
			p.Next()
			if p.Current() == mc {
				p.Next()
				continue
			}

			// Give back the character consumed since the second one wasn't
			// another quote
			p.Backtrack()
			break
		}

		if p.IsInDoubleQuotes && p.Current() == '\\' {
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

	if p.Current() != mc {
		panic("unmatched quote")
	}
	p.Next() // Consume the closing quote

	return sb.String()
}
