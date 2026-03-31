//go:build linux

package main

import (
	"os"

	"golang.org/x/term"
)

func EnableRawMode() (func(), error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	// Return a cleanup function that the caller can defer
	restore := func() {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}

	return restore, nil
}
