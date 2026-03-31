package main

import (
	"bufio"
	"os"
)

type Editor struct {
	contents  [][]rune
	cmdBuffer []rune
	mode      string
	filename  string

	r *bufio.Reader

	cursorX int
	cursorY int

	termWidth  int
	termHeight int
}

func NewEditor() *Editor {
	return &Editor{
		contents:  [][]rune{{}},
		cmdBuffer: make([]rune, 0, 64),
		mode:      "normal",
		r:         bufio.NewReader(os.Stdin),
	}
}

const ESC = 27 // ascii value for Escape

func (e *Editor) handleKey(key byte) {
	switch e.mode {
	case "normal":
		switch key {
		case 'i':
			e.mode = "insert"
		case ':':
			e.mode = "command"
			e.cmdBuffer = e.cmdBuffer[:0] // clear last command
		case 'h', 'j', 'k', 'l':
			e.moveCursor(key)
		}

	case "insert":
		switch key {
		case ESC:
			e.mode = "normal"

		case '\r', '\n':
			e.insertNewline()

		case 127, 8: // ascii backspace
			e.deleteChar()
		default:
			// check valid ascii range
			if key >= 32 && key < 127 {
				e.insertRune(rune(key))
			}
		}

	case "command":
		switch key {
		case ESC:
			e.mode = "normal"
			e.cmdBuffer = e.cmdBuffer[:0]
		case '\r', '\n':
			e.processCommand(string(e.cmdBuffer))
			e.mode = "normal"
			e.cmdBuffer = e.cmdBuffer[:0]
		case 127, 8: // Backspace
			if len(e.cmdBuffer) > 0 {
				e.cmdBuffer = e.cmdBuffer[:len(e.cmdBuffer)-1]
			}
		default:
			e.cmdBuffer = append(e.cmdBuffer, rune(key)) // append to command buffer
		}
	}
}

// TODOs
func (e *Editor) processCommand(command string) {}
func (e *Editor) moveCursor(key byte)           {}
func (e *Editor) insertNewline()                {}
func (e *Editor) deleteChar()                   {}
func (e *Editor) insertRune(r rune)             {}

func main() {}
