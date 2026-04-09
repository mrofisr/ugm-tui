//go:build linux || freebsd || openbsd || netbsd

// Package passwd parses /etc/passwd to extract system user information.
package passwd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
)

// User represents a parsed system user with group memberships.
type User struct {
	Details user.User
	Groups  []*user.Group
}

// Parse reads and parses users from the given passwd-format file.
func Parse(path string) ([]User, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	var users []User
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		u, err := parseLine(scanner.Text())
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		users = append(users, u)
	}
	_ = f.Close()

	return users, nil
}

// Load is a convenience wrapper that parses /etc/passwd.
func Load() ([]User, error) {
	return Parse("/etc/passwd")
}

func parseLine(line string) (User, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 7 {
		return User{}, errors.New("unexpected number of fields in /etc/passwd")
	}

	gecos := strings.Split(fs[4], ",")

	u := User{}
	u.Details.Uid = fs[2]
	u.Details.Gid = fs[3]
	u.Details.Username = fs[0]
	u.Details.Name = gecos[0]
	u.Details.HomeDir = fs[5]
	u.Groups = lookupGroups(u.Details)

	return u, nil
}

func lookupGroups(u user.User) []*user.Group {
	ids, _ := u.GroupIds()
	groups := make([]*user.Group, 0, len(ids))
	for _, id := range ids {
		g, _ := user.LookupGroupId(id)
		groups = append(groups, g)
	}
	return groups
}
