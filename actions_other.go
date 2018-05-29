// +build !linux

package main

func (v *View) Suspend() bool {
	messenger.Alert("Suspend is only supported on Linux")

	return false
}
