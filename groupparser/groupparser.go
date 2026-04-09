//go:build linux || freebsd || openbsd || netbsd
// +build linux freebsd openbsd netbsd

// Package groupparser parses /etc/group to extract system group information.
package groupparser

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/user"
	"strings"
)

// Group represents a parsed system group with its member users.
type Group struct {
	Details user.Group
	Users   []*user.User
}

var parsedGroups []Group

// GetGroups returns the list of parsed groups, parsing /etc/group if needed.
func GetGroups() (groups []Group) {
	if len(parsedGroups) == 0 {
		ParseGroups("/etc/group")
	}

	return parsedGroups
}

// ParseGroups reads and parses groups from the given group-format file.
func ParseGroups(path string) {
	parsedGroups = nil

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		group, err := parseLine(scanner.Text())
		if err != nil {
			_ = f.Close()
			log.Fatal(err)
		}

		parsedGroups = append(parsedGroups, group)
	}
	_ = f.Close()
}

func parseLine(line string) (Group, error) {
	fs := strings.Split(line, ":")

	if len(fs) != 4 {
		return Group{}, errors.New("unexpected number of fields in /etc/group")
	}

	group := Group{}
	group.Details.Gid = fs[2]
	group.Details.Name = fs[0]
	group.Users = parseUsers(fs[3])

	return group, nil
}

func parseUsers(groupUsernames string) []*user.User {
	usernames := strings.Split(groupUsernames, ",")
	users := make([]*user.User, 0, len(usernames))

	for _, username := range usernames {
		foundUser, _ := user.Lookup(username)
		users = append(users, foundUser)
	}

	return users
}
