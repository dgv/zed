package main

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Buffer stores the text for files that are loaded into the text editor
// It uses a rope to efficiently store the string and contains some
// simple functions for saving and wrapper functions for modifying the rope
type Buffer struct {
	// The eventhandler for undo/redo
	*EventHandler
	// This stores all the text in the buffer as an array of lines
	*LineArray

	Cursor Cursor

	// Path to the file on disk
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string

	// Whether or not the buffer has been modified since it was opened
	IsModified bool

	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time

	NumLines int
}

// The SerializedBuffer holds the types that get serialized when a buffer is saved
// These are used for the savecursor and saveundo options
type SerializedBuffer struct {
	EventHandler *EventHandler
	Cursor       Cursor
	ModTime      time.Time
}

func NewBufferFromString(text, path string) *Buffer {
	return NewBuffer(strings.NewReader(text), int64(len(text)), path)
}

// NewBuffer creates a new buffer from a given reader with a given path
func NewBuffer(reader io.Reader, size int64, path string) *Buffer {
	b := new(Buffer)
	b.LineArray = NewLineArray(size, reader)

	absPath, _ := filepath.Abs(path)

	b.Path = path
	b.AbsPath = absPath

	// The last time this file was modified
	b.ModTime, _ = GetModTime(b.Path)

	b.EventHandler = NewEventHandler(b)

	b.Update()

	// Put the cursor at the first spot
	cursorStartX := 0
	cursorStartY := 0
	// If -startpos LINE,COL was passed, use start position LINE,COL
	if len(*flagStartPos) > 0 {
		positions := strings.Split(*flagStartPos, ",")
		if len(positions) == 2 {
			lineNum, errPos1 := strconv.Atoi(positions[0])
			colNum, errPos2 := strconv.Atoi(positions[1])
			if errPos1 == nil && errPos2 == nil {
				cursorStartX = colNum
				cursorStartY = lineNum - 1
				// Check to avoid line overflow
				if cursorStartY > b.NumLines {
					cursorStartY = b.NumLines - 1
				} else if cursorStartY < 0 {
					cursorStartY = 0
				}
				// Check to avoid column overflow
				if cursorStartX > len(b.Line(cursorStartY)) {
					cursorStartX = len(b.Line(cursorStartY))
				} else if cursorStartX < 0 {
					cursorStartX = 0
				}
			}
		}
	}
	b.Cursor = Cursor{
		Loc: Loc{
			X: cursorStartX,
			Y: cursorStartY,
		},
		buf: b,
	}

	return b
}

func (b *Buffer) GetName() string {
	if b.name == "" {
		if b.Path == "" {
			return "unnamed"
		}
		return b.Path
	}
	return b.name
}

// IndentString returns a string representing one level of indentation
func (b *Buffer) IndentString() string {
	return "\t"
}

// CheckModTime makes sure that the file this buffer points to hasn't been updated
// by an external program since it was last read
// If it has, we ask the user if they would like to reload the file
func (b *Buffer) CheckModTime() {
	modTime, ok := GetModTime(b.Path)
	if ok {
		if modTime != b.ModTime {
			choice, canceled := messenger.YesNoPrompt("The file has changed since it was last read. Reload file? (y,n)")
			messenger.Reset()
			messenger.Clear()
			if !choice || canceled {
				// Don't load new changes -- do nothing
				b.ModTime, _ = GetModTime(b.Path)
			} else {
				// Load new changes
				b.ReOpen()
			}
		}
	}
}

// ReOpen reloads the current buffer from disk
func (b *Buffer) ReOpen() {
	data, err := ioutil.ReadFile(b.Path)
	txt := string(data)

	if err != nil {
		messenger.Alert(err.Error())
		return
	}
	b.EventHandler.ApplyDiff(txt)

	b.ModTime, _ = GetModTime(b.Path)
	b.IsModified = false
	b.Update()
	b.Cursor.Relocate()
}

// Update fetches the string from the rope and updates the `text` and `lines` in the buffer
func (b *Buffer) Update() {
	b.NumLines = len(b.lines)
}

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.Path)
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	//b.UpdateRules()
	dir, _ := homedir.Dir()
	str := b.String()
	data := []byte(str)
	filename = strings.Replace(filename, "~", dir, 1)
	err := ioutil.WriteFile(filename, data, 0644)
	if err == nil {
		b.Path = strings.Replace(filename, "~", dir, 1)
		b.IsModified = false
		b.ModTime, _ = GetModTime(filename)
		return err
	}
	b.ModTime, _ = GetModTime(filename)
	return err
}

func (b *Buffer) insert(pos Loc, value []byte) {
	b.IsModified = true
	b.LineArray.insert(pos, value)
	b.Update()
}
func (b *Buffer) remove(start, end Loc) string {
	b.IsModified = true
	sub := b.LineArray.remove(start, end)
	b.Update()
	return sub
}
func (b *Buffer) deleteToEnd(start Loc) {
	b.IsModified = true
	b.LineArray.DeleteToEnd(start)
	b.Update()
}

// Start returns the location of the first character in the buffer
func (b *Buffer) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (b *Buffer) End() Loc {
	return Loc{utf8.RuneCount(b.lines[b.NumLines-1].data), b.NumLines - 1}
}

// RuneAt returns the rune at a given location in the buffer
func (b *Buffer) RuneAt(loc Loc) rune {
	line := []rune(b.Line(loc.Y))
	if len(line) > 0 {
		return line[loc.X]
	}
	return '\n'
}

// Line returns a single line
func (b *Buffer) Line(n int) string {
	if n >= len(b.lines) {
		return ""
	}
	return string(b.lines[n].data)
}

func (b *Buffer) LinesNum() int {
	return len(b.lines)
}

// Lines returns an array of strings containing the lines from start to end
func (b *Buffer) Lines(start, end int) []string {
	lines := b.lines[start:end]
	var slice []string
	for _, line := range lines {
		slice = append(slice, string(line.data))
	}
	return slice
}

// Len gives the length of the buffer
func (b *Buffer) Len() int {
	return Count(b.String())
}
