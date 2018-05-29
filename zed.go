package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dgv/zed/errors"
	"github.com/dgv/zed/tcell"
)

const undoThreshold = 500 // If two events are less than n milliseconds apart, undo both of them

var (
	// The main screen
	screen tcell.Screen

	// Object to send messages and prompts to the user
	messenger *Messenger

	// The default highlighting style
	// This simply defines the default foreground and background colors
	defStyle tcell.Style

	// Version is the version number or commit hash
	// These variables should be set by the linker when compiling
	Version    = "0.0.0-unknown"
	CommitHash = "Unknown"

	// multiple views with splits
	views []*View
	// This is the current view for this tab
	mainView int

	// Event channel
	events chan tcell.Event
)

// LoadInput determines which files should be loaded into buffers
// based on the input stored in flag.Args()
func LoadInput() *Buffer {
	// There are a number of ways micro should start given its input

	// 1. If it is given a files in flag.Args(), it should open those

	// 2. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	var filename string
	var input []byte
	var err error
	args := flag.Args()
	buf := new(Buffer)

	if len(args) == 1 {
		// Option 1
		filename = args[0]
		// Check that the file exists
		var input *os.File
		if _, e := os.Stat(filename); e == nil {
			// If it exists we load it into a buffer
			input, err = os.Open(filename)
			stat, _ := input.Stat()
			defer input.Close()
			if stat.IsDir() {
				TermMessage("cannot read", filename, "because it is a directory")
				return nil
			} else if err != nil {
				TermMessage(err)
				return nil
			}
		}
		// If the file didn't exist, input will be empty, and we'll open an empty buffer
		if input != nil {
			buf = NewBuffer(input, FSize(input), filename)
		} else {
			buf = NewBufferFromString("", filename)
		}
	} else {
		// Option 2, just open an empty buffer
		buf = NewBufferFromString(string(input), filename)
	}

	return buf
}

// InitScreen creates and initializes the tcell screen
func InitScreen() {
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")

	// Initilize tcell
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Println(err)
		if err == tcell.ErrTermNotFound {
			fmt.Println("zed does not recognize your terminal:", oldTerm)
			fmt.Println("please go to https://github.com/zyedidia/mkinfo to read about how to fix this problem (it should be easy to fix).")
		}
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	screen.SetStyle(defStyle)
}

// RedrawAll redraws everything -- all the views and the messenger
func RedrawAll() {
	messenger.Clear()

	w, h := screen.Size()
	for x := 0; x < w; x++ {
		for y := 0; y < h-1; y++ {
			screen.SetContent(x, y, ' ', nil, defStyle)
		}
	}

	views[mainView].Display()
	messenger.Display()
	screen.Show()
}

// Passing -version as a flag will have micro print out the version number
var flagVersion = flag.Bool("version", false, "show the version number and information.")
var flagStartPos = flag.String("startpos", "", "LINE,COL to start the cursor at when opening a buffer.")
var flagTabSize = flag.Int("tabsize", 4, "tab size to be used")

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: zed [OPTIONS] [FILE]")
		fmt.Print("zed's options can be set via command line arguments for quick adjustments. For real configuration, please use the bindings.json file (see 'help options').\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *flagVersion {
		// If -version was passed
		fmt.Println("Version:", Version)
		fmt.Println("Commit hash:", CommitHash)
		os.Exit(0)
	}

	InitBindings()

	// Start the screen
	InitScreen()

	// This is just so if we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	// In other words we need to shut down tcell before the program crashes
	defer func() {
		if err := recover(); err != nil {
			screen.Fini()
			fmt.Println("zed encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	// Create a new messenger
	// This is used for sending the user messages in the bottom of the editor
	messenger = new(Messenger)
	messenger.style = defStyle.Bold(true)
	messenger.history = make(map[string][]string)

	// Now we load the input
	buf := LoadInput()
	if buf == nil {
		screen.Fini()
		os.Exit(1)
	}

	views = make([]*View, 1)
	views[mainView] = NewView(buf)

	events = make(chan tcell.Event, 100)

	// Here is the event loop which runs in a separate thread
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	for {
		// Display everything
		RedrawAll()

		var event tcell.Event

		// Check for new events
		select {
		case event = <-events:
		}

		for event != nil {
			switch e := event.(type) {
			case *tcell.EventResize:
				//for _, t := range tabs {
				//	t.Resize()
				//}
				views[mainView].Resize(e.Size())
			}

			if searching {
				// Since searching is done in real time, we need to redraw every time
				// there is a new event in the search bar so we need a special function
				// to run instead of the standard HandleEvent.
				HandleSearchEvent(event, views[mainView])
			} else {
				// Send it to the view
				views[mainView].HandleEvent(event)
			}

			select {
			case event = <-events:
			default:
				event = nil
			}
		}
	}
}
