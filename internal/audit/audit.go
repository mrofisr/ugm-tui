//go:build linux || freebsd || openbsd || netbsd

// Package audit provides action logging for ugm operations.
package audit

import (
	"fmt"
	"os"
	"os/user"
	"time"
)

const _logPath = "/var/log/ugm-audit.log"

var _operator = resolveOperator()

func resolveOperator() string {
	if u := os.Getenv("SUDO_USER"); u != "" {
		return u
	}
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "unknown"
}

// Log writes an action entry to the audit log file.
func Log(action, target, detail string) {
	f, err := os.OpenFile(_logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return // best-effort, don't crash TUI
	}
	defer func() { _ = f.Close() }()

	ts := time.Now().Format(time.RFC3339)
	_, _ = fmt.Fprintf(f, "%s operator=%s action=%s target=%s %s\n", ts, _operator, action, target, detail)
}
