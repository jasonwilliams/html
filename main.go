package main 

import (
    "fmt"
    "io/ioutil"
    "unicode/utf8"
    "strings"
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
    state       stateFn   // the next lexing function to enter
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
    for l.state = lexText; l.state != nil; {
        l.state = l.state(l)
    }
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
    return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

func lexText(l *lexer) stateFn {
    if strings.HasPrefix(l.input[l.pos:], string(leftBracket)) {
        fmt.Println("Inside Element")
        l.next() // step into element
        for l.accept(" >") == false { // Keep going until we hit a right bracket or a space
            l.next()
        }
        l.backup()
        // Before moving on to attributes, get the element name
        elemName := l.input[l.start + 1:l.pos]
        
        if l.next() == rightBracket {
            fmt.Printf("We have a %s element", elemName)
        }

    }

    return lexText
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