//go:build linux || freebsd || openbsd || netbsd

// Package group parses /etc/group to extract system group information.
package group

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
)

// Group represents a parsed system group with its member users.
type Group struct {
	Details user.Group
	Users   []*user.User
}

// Parse reads and parses groups from the given group-format file.
func Parse(path string) ([]Group, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	var groups []Group
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		g, err := parseLine(scanner.Text())
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		groups = append(groups, g)
	}
	_ = f.Close()

	return groups, nil
}

// Load is a convenience wrapper that parses /etc/group.
func Load() ([]Group, error) {
	return Parse("/etc/group")
}

func parseLine(line string) (Group, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 4 {
		return Group{}, errors.New("unexpected number of fields in /etc/group")
	}

	g := Group{}
	g.Details.Gid = fs[2]
	g.Details.Name = fs[0]
	g.Users = lookupUsers(fs[3])

	return g, nil
}

func lookupUsers(names string) []*user.User {
	parts := strings.Split(names, ",")
	users := make([]*user.User, 0, len(parts))
	for _, name := range parts {
		u, _ := user.Lookup(name)
		users = append(users, u)
	}
	return users
}
