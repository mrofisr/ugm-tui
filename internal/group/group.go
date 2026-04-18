//go:build linux || freebsd || openbsd || netbsd

// Package group parses /etc/group to extract system group information.
package group

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
)

// Group represents a parsed system group with its member users.
type Group struct {
	Details user.Group
	Users   []*user.User
}

// Parse reads and parses groups from the given group-format reader.
func Parse(r io.Reader) ([]Group, error) {
	var groups []Group
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		g, err := parseLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return groups, nil
}

// Load is a convenience wrapper that parses /etc/group.
func Load() ([]Group, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return nil, fmt.Errorf("open /etc/group: %w", err)
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

func parseLine(line string) (Group, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 4 {
		return Group{}, errors.New("unexpected number of fields in /etc/group")
	}

	return Group{
		Details: user.Group{
			Gid:  fs[2],
			Name: fs[0],
		},
		Users: lookupUsers(fs[3]),
	}, nil
}

func lookupUsers(names string) []*user.User {
	if names == "" {
		return nil
	}
	parts := strings.Split(names, ",")
	users := make([]*user.User, 0, len(parts))
	for _, name := range parts {
		if name == "" {
			continue
		}
		if u, err := user.Lookup(name); err == nil {
			users = append(users, u)
		}
	}
	return users
}
