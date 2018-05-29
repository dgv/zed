package main

import (
	"regexp"

	"github.com/dgv/zed/tcell"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart int

	// Is there currently a search in progress
	searching bool

	// Stores the history for searching
	searchHistory []string
)

// BeginSearch starts a search
func BeginSearch(searchStr string) {
	searchHistory = append(searchHistory, "")
	messenger.historyNum = len(searchHistory) - 1
	searching = true
	messenger.response = searchStr
	messenger.cursorx = Count(searchStr)
	messenger.Search("find what: ")
	messenger.hasPrompt = true
}

// EndSearch stops the current search
func EndSearch() {
	searchHistory[len(searchHistory)-1] = messenger.response
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
}

// exit the search mode, reset active search phrase, and clear status bar
func ExitSearch(v *View) {
	lastSearch = ""
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
	v.Cursor.ResetSelection()
}

// HandleSearchEvent takes an event and a view and will do a real time match from the messenger's output
// to the current buffer. It searches down the buffer.
func HandleSearchEvent(event tcell.Event, v *View) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEscape:
			// Exit the search mode
			ExitSearch(v)
			return
		case tcell.KeyCtrlQ, tcell.KeyCtrlC, tcell.KeyEnter:
			// Done
			EndSearch()
			return
		}
	}

	messenger.HandleEvent(event, searchHistory)

	if messenger.cursorx < 0 {
		// Done
		EndSearch()
		return
	}

	if messenger.response == "" {
		v.Cursor.ResetSelection()
		// We don't end the search though
		return
	}

	Search(messenger.response, v, true)

	v.Relocate()

	return
}

// Search searches in the view for the given regex. The down bool
// specifies whether it should search down from the searchStart position
// or up from there
func Search(searchStr string, v *View, down bool) {
	if searchStr == "" {
		return
	}
	var str string
	var charPos int
	text := v.Buf.String()
	if down {
		str = string([]rune(text)[searchStart:])
		charPos = searchStart
	} else {
		str = string([]rune(text)[:searchStart])
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return
	}
	matches := r.FindAllStringIndex(str, -1)
	var match []int
	if matches == nil {
		// Search the entire buffer now
		matches = r.FindAllStringIndex(text, -1)
		charPos = 0
		if matches == nil {
			v.Cursor.ResetSelection()
			return
		}

		if !down {
			match = matches[len(matches)-1]
		} else {
			match = matches[0]
		}
		str = text
	}

	if !down {
		match = matches[len(matches)-1]
	} else {
		match = matches[0]
	}

	if match[0] == match[1] {
		return
	}

	v.Cursor.SetSelectionStart(FromCharPos(charPos+runePos(match[0], str), v.Buf))
	v.Cursor.SetSelectionEnd(FromCharPos(charPos+runePos(match[1], str), v.Buf))
	v.Cursor.Loc = v.Cursor.CurSelection[1]
	lastSearch = searchStr
}

// Replace runs search and replace
func Replace(args []string) {
	search := string(args[0])
	replace := string(args[1])

	regex, err := regexp.Compile("(?m)" + search)
	if err != nil {
		// There was an error with the user's regex
		messenger.Alert(err.Error())
		return
	}

	view := views[mainView]

	found := 0
	all := false
	for {
		// The 'check' flag was used
		Search(search, view, true)
		if !view.Cursor.HasSelection() {
			break
		}
		view.Relocate()
		RedrawAll()
		choice, canceled := messenger.YesNoAllPrompt("perform replacement? (y, n, a, esc) ")
		if canceled {
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			return
		}
		switch choice {
		case 'y':
			view.Cursor.DeleteSelection()
			view.Buf.Insert(view.Cursor.Loc, replace)
			view.Cursor.ResetSelection()
			messenger.Reset()
			found++
		case 'n':
			if view.Cursor.HasSelection() {
				searchStart = ToCharPos(view.Cursor.CurSelection[1], view.Buf)
			} else {
				searchStart = ToCharPos(view.Cursor.Loc, view.Buf)
			}
			continue
		case 'a':
			all = true
		}
		if all {
			break
		}
	}

	if all {
		bufStr := view.Buf.String()
		matches := regex.FindAllStringIndex(bufStr, -1)
		if matches != nil && len(matches) > 0 {
			prevMatchCount := runePos(matches[0][0], bufStr)
			searchCount := runePos(matches[0][1], bufStr) - prevMatchCount
			prevMatch := matches[0]
			from := FromCharPos(prevMatch[0], view.Buf)
			to := from.Move(searchCount, view.Buf)
			adjust := Count(replace) - searchCount
			view.Buf.Replace(from, to, replace)
			if len(matches) > 1 {
				found++
				for _, match := range matches[1:] {
					found++
					matchCount := runePos(match[0], bufStr)
					searchCount = runePos(match[1], bufStr) - matchCount
					from = from.Move(matchCount-prevMatchCount+adjust, view.Buf)
					to = from.Move(searchCount, view.Buf)
					view.Buf.Replace(from, to, replace)
					prevMatch = match
					prevMatchCount = matchCount
					adjust = Count(replace) - searchCount
				}
			}
		}
		// FIXME Relocate bugs
		view.CursorEnd()
	} else {
		view.Cursor.Relocate()
	}
	RedrawAll()

	if found > 1 {
		messenger.Alert("replaced ", found, " occurrences of ", search)
	} else if found == 1 {
		messenger.Alert("replaced ", found, " occurrence of ", search)
	} else {
		messenger.Alert("nothing matched ", search)
	}
}
