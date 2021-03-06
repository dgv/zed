package main

import (
	"github.com/dgv/zed/runewidth"
	"github.com/dgv/zed/tcell"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func visualToCharPos(visualIndex int, lineN int, str string, buf *Buffer, tabsize int) (int, int, *tcell.Style) {
	charPos := 0
	var lineIdx int
	var lastWidth int
	var style *tcell.Style
	var width int
	var rw int
	for i, c := range str {
		if width >= visualIndex {
			return charPos, visualIndex - lastWidth, style
		}

		if i != 0 {
			charPos++
			lineIdx += rw
		}
		lastWidth = width
		rw = 0
		if c == '\t' {
			rw = tabsize - (lineIdx % tabsize)
			width += rw
		} else {
			rw = runewidth.RuneWidth(c)
			width += rw
		}
	}

	return -1, -1, style
}

type Char struct {
	visualLoc Loc
	realLoc   Loc
	char      rune
	// The actual character that is drawn
	// This is only different from char if it's for example hidden character
	drawChar rune
	style    tcell.Style
	width    int
}

type CellView struct {
	lines [][]*Char
}

func (c *CellView) Draw(buf *Buffer, top, height, left, width int) {
	tabsize := *flagTabSize
	indentrunes := []rune(" ")
	// if empty indentchar settings, use space
	if indentrunes == nil || len(indentrunes) == 0 {
		indentrunes = []rune(" ")
	}
	indentchar := indentrunes[0]

	c.lines = make([][]*Char, 0)

	viewLine := 0
	lineN := top

	curStyle := defStyle
	for viewLine < height {
		if lineN >= len(buf.lines) {
			break
		}

		lineStr := buf.Line(lineN)
		line := []rune(lineStr)

		colN, startOffset, startStyle := visualToCharPos(left, lineN, lineStr, buf, tabsize)
		if colN < 0 {
			colN = len(line)
		}
		viewCol := -startOffset
		if startStyle != nil {
			curStyle = *startStyle
		}

		// We'll either draw the length of the line, or the width of the screen
		// whichever is smaller
		lineLength := min(StringWidth(lineStr, tabsize), width)
		c.lines = append(c.lines, make([]*Char, lineLength))

		for viewCol < lineLength {
			if colN >= len(line) {
				break
			}

			char := line[colN]

			if viewCol >= 0 {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, char, curStyle, 1}
			}
			if char == '\t' {
				charWidth := tabsize - (viewCol+left)%tabsize
				if viewCol >= 0 {
					c.lines[viewLine][viewCol].drawChar = indentchar
					c.lines[viewLine][viewCol].width = charWidth

					indentStyle := curStyle

					c.lines[viewLine][viewCol].style = indentStyle
				}

				for i := 1; i < charWidth; i++ {
					viewCol++
					if viewCol >= 0 && viewCol < lineLength {
						c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, ' ', curStyle, 1}
					}
				}
				viewCol++
			} else if runewidth.RuneWidth(char) > 1 {
				charWidth := runewidth.RuneWidth(char)
				if viewCol >= 0 {
					c.lines[viewLine][viewCol].width = charWidth
				}
				for i := 1; i < charWidth; i++ {
					viewCol++
					if viewCol >= 0 && viewCol < lineLength {
						c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, ' ', curStyle, 1}
					}
				}
				viewCol++
			} else {
				viewCol++
			}
			colN++

		}

		// newline
		viewLine++
		lineN++
	}

	for i := top; i < top+height; i++ {
		if i >= buf.NumLines {
			break
		}
	}
}
