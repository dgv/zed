package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/dgv/clipboard"
	"github.com/dgv/zed/runewidth"
	"github.com/dgv/zed/tcell"
)

// TermMessage sends a message to the user in the terminal. This usually occurs before
// micro has been fully initialized -- ie if there is an error in the syntax highlighting
// regular expressions
// The function must be called when the screen is not initialized
// This will write the message, and wait for the user
// to press and key to continue
func TermMessage(msg ...interface{}) {
	screenWasNil := screen == nil
	if !screenWasNil {
		screen.Fini()
		screen = nil
	}

	fmt.Println(msg...)
	fmt.Print("\npress enter to continue")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	if !screenWasNil {
		InitScreen()
	}
}

// TermError sends an error to the user in the terminal. Like TermMessage except formatted
// as an error
func TermError(filename string, lineNum int, err string) {
	TermMessage(filename + ", " + strconv.Itoa(lineNum) + ": " + err)
}

// Messenger is an object that makes it easy to send messages to the user
// and get input from the user
type Messenger struct {
	log *Buffer
	// Are we currently prompting the user?
	hasPrompt bool
	// Is there a message to print
	hasMessage bool

	// Message to print
	message string
	// The user's response to a prompt
	response string
	// style to use when drawing the message
	style tcell.Style

	// We have to keep track of the cursor for prompting
	cursorx int

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	history    map[string][]string
	historyNum int
}

// Message sends a message to the user
func (m *Messenger) Search(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if m.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		m.message = displayMessage

		m.hasMessage = true
	}
}

func (m *Messenger) PromptText(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// if there is no active prompt then style and display the message as normal
	m.message = displayMessage

	m.hasMessage = true
}

// YesNoPrompt asks the user a yes or no question (waits for y or n) and returns the result
func (m *Messenger) Alert(msg ...interface{}) bool {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, msg...)
	m.hasPrompt = true
	prompt := buf.String()
	prompt += " (esc or enter) "
	m.PromptText(prompt)

	_, h := screen.Size()
	for {
		m.Clear()
		m.Display()
		screen.ShowCursor(Count(m.message), h-1)
		screen.Show()
		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape, tcell.KeyEnter:
				m.Clear()
				m.Reset()
				m.hasPrompt = false
				return true
			}
		}
	}
}

// YesNoPrompt asks the user a yes or no question (waits for y or n) and returns the result
func (m *Messenger) YesNoPrompt(prompt string) (bool, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)

	_, h := screen.Size()
	for {
		m.Clear()
		m.Display()
		screen.ShowCursor(Count(m.message), h-1)
		screen.Show()
		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyRune:
				if e.Rune() == 'y' || e.Rune() == 'Y' {
					m.hasPrompt = false
					return true, false
				} else if e.Rune() == 'n' || e.Rune() == 'N' {
					m.hasPrompt = false
					return false, false
				}
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
				m.Clear()
				m.Reset()
				m.hasPrompt = false
				return false, true
			}
		}
	}
}

// YesNoPrompt asks the user a yes or no question (waits for y or n) and returns the result
func (m *Messenger) YesNoAllPrompt(prompt string) (rune, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)

	_, h := screen.Size()
	for {
		m.Clear()
		m.Display()
		screen.ShowCursor(Count(m.message), h-1)
		screen.Show()
		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyRune:
				if e.Rune() == 'y' || e.Rune() == 'Y' {
					m.hasPrompt = false
					return 'y', false
				} else if e.Rune() == 'n' || e.Rune() == 'N' {
					m.hasPrompt = false
					return 'n', false
				} else {
					//if e.Rune() == 'a' || e.Rune() == 'A' {
					m.hasPrompt = false
					return 'a', false
				}
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
				m.Clear()
				m.Reset()
				m.hasPrompt = false
				return ' ', true
			}
		}
	}
}

// Prompt sends the user a message and waits for a response to be typed in
// This function blocks the main loop while waiting for input
func (m *Messenger) Prompt(prompt, placeholder, historyType string) (string, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)
	if _, ok := m.history[historyType]; !ok {
		m.history[historyType] = []string{""}
	} else {
		m.history[historyType] = append(m.history[historyType], "")
	}
	m.historyNum = len(m.history[historyType]) - 1

	response, canceled := placeholder, true
	m.response = response
	m.cursorx = Count(placeholder)

	RedrawAll()
	for m.hasPrompt {
		var suggestions []string
		m.Clear()

		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyCtrlQ, tcell.KeyCtrlC, tcell.KeyEscape:
				// Cancel
				m.hasPrompt = false
			case tcell.KeyEnter:
				// User is done entering their response
				m.hasPrompt = false
				response, canceled = m.response, false
				m.history[historyType][len(m.history[historyType])-1] = response
			}
		}

		m.HandleEvent(event, m.history[historyType])

		m.Clear()
		views[mainView].Display()
		m.Display()
		if len(suggestions) > 1 {
			m.DisplaySuggestions(suggestions)
		}
		screen.Show()
	}

	m.Clear()
	m.Reset()
	return response, canceled
}

// HandleEvent handles an event for the prompter
func (m *Messenger) HandleEvent(event tcell.Event, history []string) {
	switch e := event.(type) {
	case *tcell.EventKey:
		if e.Key() != tcell.KeyRune || e.Modifiers() != 0 {
			for key, actions := range bindings {
				if e.Key() == key.keyCode {
					if e.Key() == tcell.KeyRune {
						if e.Rune() != key.r {
							continue
						}
					}
					if e.Modifiers() == key.modifiers {
						for _, action := range actions {
							funcName := FuncName(action)
							switch funcName {
							case "main.(*View).CursorUp":
								if m.historyNum > 0 {
									m.historyNum--
									m.response = history[m.historyNum]
									m.cursorx = Count(m.response)
								}
							case "main.(*View).CursorDown":
								if m.historyNum < len(history)-1 {
									m.historyNum++
									m.response = history[m.historyNum]
									m.cursorx = Count(m.response)
								}
							case "main.(*View).CursorLeft":
								if m.cursorx > 0 {
									m.cursorx--
								}
							case "main.(*View).CursorRight":
								if m.cursorx < Count(m.response) {
									m.cursorx++
								}
							case "main.(*View).CursorStart", "main.(*View).StartOfLine":
								m.cursorx = 0
							case "main.(*View).CursorEnd", "main.(*View).EndOfLine":
								m.cursorx = Count(m.response)
							case "main.(*View).Backspace":
								if m.cursorx > 0 {
									m.response = string([]rune(m.response)[:m.cursorx-1]) + string([]rune(m.response)[m.cursorx:])
									m.cursorx--
								}
							case "main.(*View).Paste":
								clip, _ := clipboard.ReadAll()
								m.response = Insert(m.response, m.cursorx, clip)
								m.cursorx += Count(clip)
							}
						}
					}
				}
			}
		}
		switch e.Key() {
		case tcell.KeyRune:
			m.response = Insert(m.response, m.cursorx, string(e.Rune()))
			m.cursorx++
		}
		history[m.historyNum] = m.response

	case *tcell.EventPaste:
		clip := e.Text()
		m.response = Insert(m.response, m.cursorx, clip)
		m.cursorx += Count(clip)
	}
}

// Reset resets the messenger's cursor, message and response
func (m *Messenger) Reset() {
	m.cursorx = 0
	m.message = ""
	m.response = ""
	m.hasMessage = false
	m.hasPrompt = false
}

// Clear clears the line at the bottom of the editor
func (m *Messenger) Clear() {
	w, h := screen.Size()
	for x := 0; x < w; x++ {
		screen.SetContent(x, h-1, ' ', nil, m.style)
	}
}

func (m *Messenger) DisplaySuggestions(suggestions []string) {
	w, screenH := screen.Size()

	y := screenH - 2

	for x := 0; x < w; x++ {
		screen.SetContent(x, y, ' ', nil, m.style)
	}

	x := 0
	for _, suggestion := range suggestions {
		for _, c := range suggestion {
			screen.SetContent(x, y, c, nil, m.style)
			x++
		}
		screen.SetContent(x, y, ' ', nil, m.style)
		x++
	}
}

func (m *Messenger) Display() {
	w, h := screen.Size()
	if m.hasMessage {
		//if m.hasPrompt {
		m.Clear()
		runes := []rune(m.message + m.response)
		posx := 0
		for x := 0; x < len(runes); x++ {
			screen.SetContent(posx, h-1, runes[x], nil, m.style)
			posx += runewidth.RuneWidth(runes[x])
		}
		//}
	}

	if m.hasPrompt {
		screen.ShowCursor(Count(m.message)+m.cursorx, h-1)
		screen.Show()
	}

	if !m.hasMessage && !m.hasPrompt {
		h = h - 1
		v := views[mainView]
		modified := " "
		if v.Buf.IsModified {
			modified = "*"
		}
		runes := []rune(fmt.Sprintf(" %s%s (%d,%d)", modified, path.Base(v.Buf.GetName()), v.Cursor.Y+1, v.Cursor.GetVisualX()+1))
		for x := 0; x < len(runes); x++ {
			screen.SetContent(x, h, runes[x], nil, m.style)
		}
		screen.SetContent(w-len(runes), h, ' ', nil, m.style)
	}
}
