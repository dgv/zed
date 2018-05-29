package main

import (
	"os"
	"strings"
	"time"

	"github.com/dgv/zed/tcell"
)

// The View struct stores information about a view into a buffer.
// It stores information about the cursor, and the viewport
// that the user sees the buffer from.
type View struct {
	// A pointer to the buffer's cursor for ease of access
	Cursor *Cursor

	// The topmost line, used for vertical scrolling
	Topline int
	// The leftmost column, used for horizontal scrolling
	leftCol int

	// Actual width and height
	Width  int
	Height int

	// Where this view is located
	x, y int

	// How much to offset because of line numbers
	lineNumOffset int

	// The buffer
	Buf *Buffer

	// lastCutTime stores when the last ctrl+k was issued.
	// It is used for clearing the clipboard to replace it with fresh cut lines.
	lastCutTime time.Time

	// freshClip returns true if the clipboard has never been pasted.
	freshClip bool

	cellview *CellView
}

// NewView returns a new fullscreen view
func NewView(buf *Buffer) *View {
	screenW, screenH := screen.Size()
	return NewViewWidthHeight(buf, screenW, screenH)
}

// NewViewWidthHeight returns a new view with the specified width and height
// Note that w and h are raw column and row values
func NewViewWidthHeight(buf *Buffer, w, h int) *View {
	v := new(View)

	v.Resize(screen.Size())
	v.x, v.y = 0, 0

	v.Width = w
	v.Height = h - 1
	v.cellview = new(CellView)

	v.OpenBuffer(buf)

	return v
}

func (v *View) Resize(w, h int) {
	v.Width = w
	v.Height = h - 1
}

func (v *View) paste(clip string) {
	leadingWS := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	clip = strings.Replace(clip, "\n", "\n"+leadingWS, -1)
	v.Buf.Insert(v.Cursor.Loc, clip)
	v.Cursor.Loc = v.Cursor.Loc.Move(Count(clip), v.Buf)
	v.freshClip = false
}

// ScrollUp scrolls the view up n lines (if possible)
func (v *View) ScrollUp(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline-n >= 0 {
		v.Topline -= n
	} else if v.Topline > 0 {
		v.Topline--
	}
}

// ScrollDown scrolls the view down n lines (if possible)
func (v *View) ScrollDown(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline+n <= v.Buf.NumLines {
		v.Topline += n
	} else if v.Topline < v.Buf.NumLines-1 {
		v.Topline++
	}
}

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
func (v *View) CanClose() bool {
	if v.Buf.IsModified {
		var choice bool
		var canceled bool
		choice, canceled = messenger.YesNoPrompt("do you want to save the changes ? (y, n or esc) ")
		if !canceled {
			if choice {
				v.Save()
				return true
			} else {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

// OpenBuffer opens a new buffer in this view.
// This resets the topline, event handler and cursor.
func (v *View) OpenBuffer(buf *Buffer) {
	screen.Clear()
	v.Buf = buf
	v.Cursor = &buf.Cursor
	v.Topline = 0
	v.leftCol = 0
	v.Cursor.ResetSelection()
	v.Relocate()
	v.Center()
}

// Open opens the given file in the view
func (v *View) Open(filename string) {
	home, _ := homedir.Dir()
	filename = strings.Replace(filename, "~", home, 1)
	file, err := os.Open(filename)
	fileInfo, _ := os.Stat(filename)

	if err == nil && fileInfo.IsDir() {
		messenger.Alert(filename, " is a directory")
		return
	}

	defer file.Close()

	var buf *Buffer
	if err != nil {
		_, cancel := messenger.Prompt(err.Error()+" (enter to create it or esc to cancel)", "", "")
		// File does not exist -- create an empty buffer with that name
		if cancel {
			return
		}
		buf = NewBufferFromString("", filename)
	} else {
		buf = NewBuffer(file, FSize(file), filename)
	}
	v.OpenBuffer(buf)
}

// ReOpen reloads the current buffer
func (v *View) ReOpen() {
	if v.CanClose() {
		screen.Clear()
		v.Buf.ReOpen()
		v.Relocate()
	}
}

func (v *View) Bottomline() int {
	screenX, screenY := 0, 0
	numLines := 0
	for lineN := v.Topline; lineN < v.Topline+v.Height; lineN++ {
		line := v.Buf.Line(lineN)

		colN := 0
		for _, ch := range line {
			if screenX >= v.Width-v.lineNumOffset {
				screenX = 0
				screenY++
			}

			if ch == '\t' {
				screenX += *flagTabSize - 1
			}

			screenX++
			colN++
		}
		screenX = 0
		screenY++
		numLines++

		if screenY >= v.Height {
			break
		}
	}
	return numLines + v.Topline
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate() bool {
	height := v.Bottomline() - v.Topline
	ret := false
	cy := v.Cursor.Y
	scrollmargin := 0
	if cy < v.Topline+scrollmargin && cy > scrollmargin-1 {
		v.Topline = cy - scrollmargin
		ret = true
	} else if cy < v.Topline {
		v.Topline = cy
		ret = true
	}
	if cy > v.Topline+height-1-scrollmargin && cy < v.Buf.NumLines-scrollmargin {
		v.Topline = cy - height + 1 + scrollmargin
		ret = true
	} else if cy >= v.Buf.NumLines-scrollmargin && cy > height {
		v.Topline = v.Buf.NumLines - height
		ret = true
	}

	cx := v.Cursor.GetVisualX()
	if cx < v.leftCol {
		v.leftCol = cx
		ret = true
	}
	if cx+v.lineNumOffset+1 > v.leftCol+v.Width {
		v.leftCol = cx - v.Width + v.lineNumOffset + 1
		ret = true
	}

	return ret
}

func (v *View) ExecuteActions(actions []func(*View) bool) bool {
	relocate := false
	//readonlyBindingsList := []string{"Delete", "Insert", "Backspace", "Cut", "Play", "Paste", "Move", "Add", "DuplicateLine", "Macro"}
	for _, action := range actions {
		readonlyBindingsResult := false
		if !readonlyBindingsResult {
			// call the key binding
			relocate = action(v) || relocate
		}
	}

	return relocate
}

// HandleEvent handles an event passed by the main loop
func (v *View) HandleEvent(event tcell.Event) {
	// This bool determines whether the view is relocated at the end of the function
	// By default it's true because most events should cause a relocate
	relocate := true

	v.Buf.CheckModTime()

	switch e := event.(type) {
	case *tcell.EventKey:
		// Check first if input is a key binding, if it is we 'eat' the input and don't insert a rune
		isBinding := false
		for key, actions := range bindings {
			if e.Key() == key.keyCode {
				if e.Key() == tcell.KeyRune {
					if e.Rune() != key.r {
						continue
					}
				}
				if e.Modifiers() == key.modifiers {
					relocate = false
					isBinding = true
					relocate = v.ExecuteActions(actions)
					break
				}
			}
		}
		if !isBinding && e.Key() == tcell.KeyRune {
			// Check viewtype if readonly don't insert a rune (readonly help and log view etc.)
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.Buf.Insert(v.Cursor.Loc, string(e.Rune()))
			v.Cursor.Right()
		}
	case *tcell.EventPaste:
		v.paste(e.Text())

	}

	if relocate {
		v.Relocate()
		// We run relocate again because there's a bug with relocating with softwrap
		// when for example you jump to the bottom of the buffer and it tries to
		// calculate where to put the topline so that the bottom line is at the bottom
		// of the terminal and it runs into problems with visual lines vs real lines.
		// This is (hopefully) a temporary solution
		//v.Relocate()
	}
}

func (v *View) DisplayView() {
	xOffset := v.x + v.lineNumOffset
	yOffset := v.y

	height := v.Height
	width := v.Width
	left := v.leftCol
	top := v.Topline

	v.cellview.Draw(v.Buf, top, height, left, width-v.lineNumOffset)

	//screenX := v.x
	realLineN := top - 1
	visualLineN := 0
	var line []*Char
	for visualLineN, line = range v.cellview.lines {
		var firstChar *Char
		if len(line) > 0 {
			firstChar = line[0]
		}

		//var softwrapped bool
		if firstChar != nil {
			realLineN = firstChar.realLoc.Y
		} else {
			realLineN++
		}

		var lastChar *Char
		cursorSet := false
		for _, char := range line {
			if char != nil {
				lineStyle := char.style

				charLoc := char.realLoc
				if v.Cursor.HasSelection() &&
					(charLoc.GreaterEqual(v.Cursor.CurSelection[0]) && charLoc.LessThan(v.Cursor.CurSelection[1]) ||
						charLoc.LessThan(v.Cursor.CurSelection[0]) && charLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
					// The current character is selected
					lineStyle = defStyle.Reverse(true)
				}

				if !v.Cursor.HasSelection() &&
					v.Cursor.Y == char.realLoc.Y && v.Cursor.X == char.realLoc.X && !cursorSet {
					screen.ShowCursor(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y)
					cursorSet = true
				}

				screen.SetContent(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y, char.drawChar, nil, lineStyle)

				lastChar = char
			}
		}

		lastX := 0
		var realLoc Loc
		var visualLoc Loc
		if lastChar != nil {
			lastX = xOffset + lastChar.visualLoc.X + lastChar.width
			if !v.Cursor.HasSelection() &&
				v.Cursor.Y == lastChar.realLoc.Y && v.Cursor.X == lastChar.realLoc.X+1 {
				screen.ShowCursor(lastX, yOffset+lastChar.visualLoc.Y)
			}
			realLoc = Loc{lastChar.realLoc.X + 1, realLineN}
			visualLoc = Loc{lastX - xOffset, lastChar.visualLoc.Y}
		} else if len(line) == 0 {
			if !v.Cursor.HasSelection() &&
				v.Cursor.Y == realLineN {
				screen.ShowCursor(xOffset, yOffset+visualLineN)
			}
			lastX = xOffset
			realLoc = Loc{0, realLineN}
			visualLoc = Loc{0, visualLineN}
		}

		if v.Cursor.HasSelection() &&
			(realLoc.GreaterEqual(v.Cursor.CurSelection[0]) && realLoc.LessThan(v.Cursor.CurSelection[1]) ||
				realLoc.LessThan(v.Cursor.CurSelection[0]) && realLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
			// The current character is selected
			selectStyle := defStyle.Reverse(true)
			screen.SetContent(xOffset+visualLoc.X, yOffset+visualLoc.Y, ' ', nil, selectStyle)
		}
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	v.DisplayView()
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if v.Cursor.Y-v.Topline < 0 || v.Cursor.Y-v.Topline > v.Height-1 || v.Cursor.HasSelection() {
		screen.HideCursor()
	}

}
