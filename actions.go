package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgv/clipboard"
)

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		return true
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center() bool {
	v.Topline = v.Cursor.Y - v.Height/2
	if v.Topline+v.Height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.Height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp() bool {
	v.deselect(0)
	v.Cursor.Up()

	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown() bool {
	v.deselect(1)
	v.Cursor.Down()

	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}

	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}

	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight() bool {
	v.Cursor.WordRight()

	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft() bool {
	v.Cursor.WordLeft()

	return true
}

// SelectUp selects up one line
func (v *View) SelectUp() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// SelectDown selects down one line
func (v *View) SelectDown() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft() bool {
	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight() bool {
	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine() bool {
	v.deselect(0)

	v.Cursor.Start()

	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine() bool {
	v.deselect(0)

	v.Cursor.End()

	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart() bool {
	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0

	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd() bool {
	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()

	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart()
	v.Cursor.SelectTo(v.Buf.Start())

	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd()
	v.Cursor.SelectTo(v.Buf.End())

	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()

	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline() bool {
	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	v.Buf.Insert(v.Cursor.Loc, "\n")
	v.Cursor.Right()
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	return true
}

// Backspace deletes the previous character
func (v *View) Backspace() bool {
	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc.GreaterThan(v.Buf.Start()) {
		v.Cursor.Left()
		cx, cy := v.Cursor.X, v.Cursor.Y
		v.Cursor.Right()
		loc := v.Cursor.Loc
		v.Buf.Remove(loc.Move(-1, v.Buf), loc)
		v.Cursor.X, v.Cursor.Y = cx, cy
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	return true
}

// Delete deletes the next character
func (v *View) Delete() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}

	return true
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab() bool {
	if v.Cursor.HasSelection() {
		return false
	}

	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	for i := 0; i < bytesUntilIndent; i++ {
		v.Cursor.Right()
	}

	return true
}

// Save the buffer to disk
func (v *View) Save() bool {
	// If this is an empty buffer, ask for a filename
	if v.Buf.Path == "" {
		v.SaveAs()
	} else {
		v.saveToFile(v.Buf.Path)
	}

	return false
}

// This function saves the buffer to `filename` and changes the buffer's path and name
// to `filename` if the save is successful
func (v *View) saveToFile(filename string) {
	err := v.Buf.SaveAs(filename)
	if err != nil {
		messenger.Alert(strings.Replace(err.Error(), "open", "save", 1))
	} else {
		v.Buf.Path = filename
		v.Buf.name = filename
	}
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs() bool {
	filename, canceled := messenger.Prompt("save as: ", "", "Save")
	if !canceled {
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")
		v.saveToFile(filename)
	}

	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find() bool {
	searchStr := ""
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStr = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch(searchStr)

	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	if lastSearch == "" {
		return true
	}
	Search(lastSearch, v, true)

	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	Search(lastSearch, v, false)

	return true
}

func (v *View) Replace() bool {
	searchStr := ""
	if v.Cursor.HasSelection() {
		searchStr = v.Cursor.GetSelection()
	}
	search, canceled := messenger.Prompt("find what: ", searchStr, "Replace")
	if !canceled && len(search) > 0 {
		replace, canceled := messenger.Prompt("replace with: ", "", "")
		if !canceled {
			Replace([]string{search, replace})
		}
	}
	return true
}

// Undo undoes the last action
func (v *View) Undo() bool {
	v.Buf.Undo()

	return true
}

// Redo redoes the last action
func (v *View) Redo() bool {
	v.Buf.Redo()

	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.CopySelection("clipboard")
		v.freshClip = true
	}

	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine() bool {
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	/*
		if v.freshClip == true {
			if v.Cursor.HasSelection() {
				if clip, err := clipboard.ReadAll(); err != nil {
					messenger.Error(err)
				} else {
					clipboard.WriteAll(clip + v.Cursor.GetSelection())
				}
			}
		} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
			v.Copy()
		}
		v.freshClip = true
	*/
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()

	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.CopySelection("clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true

		return true
	}

	return false
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine() bool {
	if v.Cursor.HasSelection() {
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		v.Cursor.Right()
	}

	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine() bool {
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()

	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() bool {
	clip, _ := clipboard.ReadAll()
	v.paste(clip)

	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() bool {
	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile() bool {
	if v.CanClose() {
		input, canceled := messenger.Prompt("open: ", "", "")
		if !canceled {
			filename := strings.Join(SplitCommandArgs(input), " ")
			views[mainView].Open(filename)
		}
	}

	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start() bool {
	v.Topline = 0

	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End() bool {
	if v.Height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.Height
	}

	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp() bool {
	if v.Topline > v.Height {
		v.ScrollUp(v.Height)
	} else {
		v.Topline = 0
	}

	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown() bool {
	if v.Buf.NumLines-(v.Topline+v.Height) > v.Height {
		v.ScrollDown(v.Height)
	} else if v.Buf.NumLines >= v.Height {
		v.Topline = v.Buf.NumLines - v.Height
	}

	return false
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp() bool {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.Height)

	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown() bool {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.Height)

	return true
}

// GotoLine jumps to a line and moves the view accordingly.
func (v *View) GotoLine() bool {
	// Prompt for line number
	linestring, canceled := messenger.Prompt("go to line: ", "", "LineNumber")
	if canceled {
		return false
	}
	lineint, err := strconv.Atoi(linestring)
	lineint = lineint - 1 // fix offset
	if err != nil {
		messenger.Alert(err) // return errors
		return false
	}
	// Move cursor and view if possible.
	if lineint < v.Buf.NumLines && lineint >= 0 {
		v.Cursor.X = 0
		v.Cursor.Y = lineint

		return true
	}
	messenger.Alert("only ", v.Buf.NumLines, " lines to jump")
	return false
}

// Escape leaves current mode
func (v *View) Escape() bool {
	// check if user is searching, or the last search is still active
	if searching || lastSearch != "" {
		ExitSearch(v)
		return true
	}
	// check if a prompt is shown, hide it and don't quit
	if messenger.hasPrompt {
		messenger.Reset()
		return true
	}

	return false
}

// Quit this will close the current tab or view that is open
func (v *View) Quit() bool {
	// Make sure not to quit if there are unsaved changes
	if v.CanClose() {
		screen.Fini()
		os.Exit(0)
	}

	return false
}

// None is no action
func None() bool {
	return false
}
