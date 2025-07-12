package main

import (
	"golang.design/x/clipboard"
)

func main() {
	// We need the clipboard in order to be able to copy the password
	// GIMME ALL YOUR CLIPBOARDS
	// ALL YOUR THINGS AND PASSWORDS TOO!
	// Lyrics taken from 'Gimme All Your Clipboard' by 'ZZ TOP'
	err := clipboard.Init()
	checkError(err)

	setupHeadless()
}
