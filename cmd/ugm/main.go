// The ugm command is a TUI to view and manage UNIX users and groups.
package main

import (
	"fmt"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"

	"github.com/mrofisr/ugm-tui/internal/tui"
)

// version is set via ldflags at build time (e.g. -ldflags "-X main.version=1.0.0").
var version = "dev"

var _supportedOS = map[string]bool{
	"linux":   true,
	"freebsd": true,
	"openbsd": true,
	"netbsd":  true,
}

const _usage = `ugm — a terminal UI to view and manage UNIX users and groups

Usage:
  sudo ugm

Flags:
  -h, --help      Show this help message
  -v, --version   Show version

Documentation: https://github.com/mrofisr/ugm-tui`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			fmt.Println(_usage)
			return
		case "-v", "--version", "version":
			fmt.Printf("ugm %s\n", version)
			return
		}
	}

	if !_supportedOS[runtime.GOOS] {
		fmt.Fprintf(os.Stderr, "ugm: unsupported OS %q. Supported: linux, freebsd, openbsd, netbsd.\n", runtime.GOOS)
		os.Exit(1)
	}

	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "ugm: must be run as root.\n\nUsage:\n  sudo ugm\n\nRun 'ugm --help' for more information.")
		os.Exit(1)
	}

	m, err := tui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ugm: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ugm: %v\n", err)
		os.Exit(1)
	}
}
