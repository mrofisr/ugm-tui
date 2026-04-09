// The ugm command is a TUI to view and manage UNIX users and groups.
package main

import (
	"fmt"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"

	"github.com/mrofisr/ugm-tui/internal/tui"
)

var _supportedOS = map[string]bool{
	"linux":   true,
	"freebsd": true,
	"openbsd": true,
	"netbsd":  true,
}

func main() {
	if !_supportedOS[runtime.GOOS] {
		fmt.Println("Current OS not supported. Refer to the documentation for more information.")
		os.Exit(0)
	}

	if os.Geteuid() != 0 {
		fmt.Println("ugm must be run as root. Use: sudo ugm")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
