package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printBanner()
		return
	}

	switch os.Args[1] {
	case "status":
		printStatusStub()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Printf("kickdesk: development command center (%s)\n", version)
}

func printStatusStub() {
	fmt.Println("(no apps configured yet — add ~/.config/kickdesk/config.json)")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: kickdesk [status]\n")
}
