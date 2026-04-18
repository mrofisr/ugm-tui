//go:build linux || freebsd || openbsd || netbsd

// Package passwd parses /etc/passwd to extract system user information.
package passwd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
)

// User represents a parsed system user with group memberships.
type User struct {
	Details user.User
	Shell   string
	Groups  []*user.Group
}

// Parse reads and parses users from the given passwd-format reader.
func Parse(r io.Reader) ([]User, error) {
	var users []User
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		u, err := parseLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return users, nil
}

// Load is a convenience wrapper that parses /etc/passwd.
func Load() ([]User, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("open /etc/passwd: %w", err)
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

func parseLine(line string) (User, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 7 {
		return User{}, errors.New("unexpected number of fields in /etc/passwd")
	}

	gecos, _, _ := strings.Cut(fs[4], ",")

	return User{
		Details: user.User{
			Uid:      fs[2],
			Gid:      fs[3],
			Username: fs[0],
			Name:     gecos,
			HomeDir:  fs[5],
		},
		Shell:  fs[6],
		Groups: lookupGroups(fs[2]),
	}, nil
}

func lookupGroups(uid string) []*user.Group {
	u, err := user.LookupId(uid)
	if err != nil {
		return nil
	}
	ids, err := u.GroupIds()
	if err != nil {
		return nil
	}
	groups := make([]*user.Group, 0, len(ids))
	for _, id := range ids {
		if g, err := user.LookupGroupId(id); err == nil {
			groups = append(groups, g)
		}
	}
	return groups
}
