package main

import "syscall"

// Suspend sends micro to the background. This is the same as pressing CtrlZ in most unix programs.
// This only works on linux and has no default binding.
// This code was adapted from the suspend code in nsf/godit
func (v *View) Suspend() bool {
	screenWasNil := screen == nil

	if !screenWasNil {
		screen.Fini()
		screen = nil
	}

	// suspend the process
	pid := syscall.Getpid()
	tid := syscall.Gettid()
	err := syscall.Tgkill(pid, tid, syscall.SIGSTOP)
	if err != nil {
		TermMessage(err)
	}

	if !screenWasNil {
		InitScreen()
	}

	return true
}
