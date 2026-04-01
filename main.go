package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
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

	message string // status messages
	quit    bool
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

func (e *Editor) processCommand(cmdLine string) {
	cmdLine = strings.TrimSpace(cmdLine)
	if cmdLine == "" {
		return
	}

	parts := strings.Fields(cmdLine) // split on whitespace
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}

	switch cmd {
	case "w", "write":
		filename := e.filename
		if arg != "" {
			filename = arg
		}

		if filename == "" {
			e.setMessage("No filename given. Use :w <filename>")
			return
		}

		if e.saveToFile(filename) {
			e.filename = filename
			e.setMessage(fmt.Sprintf("\"%s\" written", filename))
		}

	case "q", "quit":
		// TODO: add check for unsaved changes
		e.quit = true

	case "wq", "x":
		filename := e.filename
		if arg != "" {
			filename = arg
		}

		if filename == "" {
			e.setMessage("No filename given")
			return
		}

		if e.saveToFile(filename) {
			e.filename = filename
			e.quit = true
		}

	default:
		e.setMessage(fmt.Sprintf("Unknown command: %s", cmd))
	}
}

func (e *Editor) saveToFile(filename string) bool {
	if filename == "" {
		e.setMessage("No filename specified")
		return false
	}

	file, err := os.Create(filename)
	if err != nil {
		e.setMessage(fmt.Sprintf("Error creating file: %v", err))
		return false
	}
	defer file.Close()

	for _, line := range e.contents {
		_, err := file.WriteString(string(line) + "\n")
		if err != nil {
			e.setMessage(fmt.Sprintf("Error writing file: %v", err))
			return false
		}
	}

	return true
}

func (e *Editor) refreshScreen() {
	fmt.Print("\x1b[?25l")
	fmt.Print("\x1b[2J\x1b[H")

	for i, line := range e.contents {
		if i >= e.termHeight-1 {
			break
		}
		fmt.Print(string(line) + "\r\n")
	}

	// status / message line
	fmt.Printf("\x1b[%d;1H", e.termHeight)
	if e.message != "" {
		fmt.Print(e.message)
		e.message = "" // clear after showing
	} else if e.mode == "insert" {
		fmt.Print(" -- INSERT -- ")
	} else if e.mode == "command" {
		fmt.Print("Command: " + string(e.cmdBuffer))
	} else {
		fmt.Print(" NORMAL ")
	}

	// THIS must be last
	fmt.Printf("\x1b[%d;%dH", e.cursorY+1, e.cursorX+1)
	fmt.Print("\x1b[?25h")
}

// helper to set status message
func (e *Editor) setMessage(msg string) {
	e.message = msg
}

func (e *Editor) drawStatus() {
	fmt.Printf("\x1b[%d;1H", e.termHeight) // move to bottom line
	if e.message != "" {
		fmt.Print(e.message)
	}
}

func (e *Editor) getTermSize() {
	var fd = os.Stdin.Fd()
	if term.IsTerminal(int(fd)) {
		width, height, err := term.GetSize(int(fd))
		e.termWidth = width
		e.termHeight = height
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting terminal size: %v\n", err)
		}
	}

}

func (e *Editor) moveCursor(key byte) {
	switch key {
	case 'h':
		if e.cursorX > 0 {
			e.cursorX--
		}
	case 'j':
		if e.cursorY < len(e.contents)-1 {
			e.cursorY++
			if e.cursorX > len(e.contents[e.cursorY]) {
				e.cursorX = len(e.contents[e.cursorY])
			}
		}
	case 'k':
		if e.cursorY > 0 {
			e.cursorY--
			if e.cursorX > len(e.contents[e.cursorY]) {
				e.cursorX = len(e.contents[e.cursorY])
			}
		}
	case 'l':
		if e.cursorX < len(e.contents[e.cursorY]) {
			e.cursorX++
		}
	}
}

func (e *Editor) insertNewline() {
	before := append([]rune{}, e.contents[e.cursorY][:e.cursorX]...)
	after := append([]rune{}, e.contents[e.cursorY][e.cursorX:]...)

	// split line
	e.contents[e.cursorY] = before

	e.contents = append(e.contents, nil)
	copy(e.contents[e.cursorY+2:], e.contents[e.cursorY+1:]) // shift lines down

	// split line
	e.contents[e.cursorY+1] = after

	e.cursorY++
	e.cursorX = 0
}

func (e *Editor) deleteChar() {
	if e.cursorX > 0 {
		e.contents[e.cursorY] = append(
			e.contents[e.cursorY][:e.cursorX-1],
			e.contents[e.cursorY][e.cursorX:]...,
		)
		e.cursorX--
		return
	}

	if e.cursorX == 0 && e.cursorY > 0 {
		prevLineLen := len(e.contents[e.cursorY-1])

		e.contents[e.cursorY-1] = append(
			e.contents[e.cursorY-1],
			e.contents[e.cursorY]...,
		)

		e.contents = append(e.contents[:e.cursorY], e.contents[e.cursorY+1:]...)

		e.cursorY--
		e.cursorX = prevLineLen
		return
	}
}

func (e *Editor) insertRune(r rune) {
	e.contents[e.cursorY] = append(
		e.contents[e.cursorY][:e.cursorX],
		append([]rune{r}, e.contents[e.cursorY][e.cursorX:]...)...,
	)

	e.cursorX++
}

func main() {
	e := NewEditor()
	e.getTermSize()

	restore, err := EnableRawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error enabling raw mode: %v\n", err)
		return
	}
	defer restore()

	for {
		e.refreshScreen()
		//e.drawStatus()

		key, err := e.r.ReadByte()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			break
		}

		e.handleKey(key)

		if e.quit {
			break
		}
	}
}
