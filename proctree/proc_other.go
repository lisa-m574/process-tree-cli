//go:build !linux

package proctree

import "errors"

// ReadProcesses has no implementation outside Linux, since there's no
// /proc to read. BuildTree and Fprint still work fine on any platform
// if you supply your own []Process.
func ReadProcesses() ([]Process, error) {
	return nil, errors.New("proctree: reading live processes is only supported on linux")
}
