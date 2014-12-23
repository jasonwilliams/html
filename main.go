package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"unicode/utf8"
)

type NodeType uint32

const (
	ErrorNode NodeType = iota
	TextNode
	DocumentNode
	ElementNode
	CommentNode
	DoctypeNode
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// our lexer will hold the current state of the scanner
type lexer struct {
	name        string
	input       string
	pos         int
	start       int
	width       int
	lastPos     int
	items       chan Node
	parentDepth int
	state       stateFn // the next lexing function to enter
	indent      int
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width

	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	l.indent = 0
	for l.state = lexElement; l.state != nil; {
		l.state = l.state(l)
		time.Sleep(time.Second * 1)
	}
}

func (l *lexer) getIndent() string {
	switch l.indent {
	case 0:
		return ""
	case 1:
		return "    "
	default:
		indentString := ""
		for i := 0; i < l.indent; i++ {
			indentString += "    "
		}
		return indentString
	}
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

func lexElement(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], string(leftBracket)) {
		/* parse closing elements */
		if strings.HasPrefix(l.input[l.pos+1:], string(rightSlash)) {
			fmt.Println("in closing element")
			return lexClosingElement

		}
		startingPos := l.pos
		for l.accept(" >") == false { // Keep going until we hit a right bracket or a space
			l.next()
		}
		l.backup()

		// Before moving on to attributes, get the element name
		elemName := l.input[startingPos+1 : l.pos]
		if l.input[l.pos] == rightBracket {
			fmt.Printf("%s<%s> \n", l.getIndent(), elemName)
			return lexElement
		} else if l.input[l.pos] == ' ' {
			return lexAttribute
		}
	} else if strings.HasPrefix(l.input[l.pos:], string(rightBracket)) {
		fmt.Printf("%q\n", l.peek())
		fmt.Println("We have reached the end of the element, start parsing the value")
		l.next()
		return lexValue
	} else {
		fmt.Printf("%q\n", l.input[l.pos])
		fmt.Println("fell into else")
		l.next()
	}

	l.indent++
	return lexElement
}

func lexValue(l *lexer) stateFn {
	for l.accept("<") == false { // Keep going until we hit a right bracket or a space
		l.next()
	}
	l.backup()
	fmt.Println("Hit a new element")
	return lexElement
}

func lexClosingElement(l *lexer) stateFn {
	/* At the start of a closing element */
	if strings.HasPrefix(l.input[l.pos:], string(leftBracket)) {
		fmt.Println("inside the start of a closing element")
		fmt.Printf("%q", l.peek())
	}

	return lexElement
}

func lexAttribute(l *lexer) stateFn {
	fmt.Println("now in attribute parsing")
	return lexAttribute
}

func main() {
	data, err := ioutil.ReadFile("test/test.html")
	if err != nil {
		fmt.Println("oops, cannot find input file")
	}

	lex := &lexer{input: string(data)}
	fmt.Println("running lexer")
	lex.run()

}
