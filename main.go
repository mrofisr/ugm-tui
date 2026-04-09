// Package main is the entry point for the ugm TUI application.
package main

import (
	"fmt"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/ariasmn/ugm/internal/tui"
)

var supportedOS = map[string]bool{
	"linux":   true,
	"freebsd": true,
	"openbsd": true,
	"netbsd":  true,
}

func main() {
	if !supportedOS[runtime.GOOS] {
		fmt.Println("Current OS not supported. Refer to the documentation for more information.")
		os.Exit(0)
	}

	if os.Geteuid() != 0 {
		fmt.Println("ugm must be run as root. Use: sudo ugm")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.InitialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
