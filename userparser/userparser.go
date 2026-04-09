//go:build linux || freebsd || openbsd || netbsd
// +build linux freebsd openbsd netbsd

// Package userparser parses /etc/passwd to extract system user information.
package userparser

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/user"
	"strings"
)

// User represents a parsed system user with group memberships.
type User struct {
	Details user.User
	Groups  []*user.Group
}

var parsedUsers []User

// GetUsers returns the list of parsed users, parsing /etc/passwd if needed.
func GetUsers() (users []User) {
	if len(parsedUsers) == 0 {
		ParseUsers("/etc/passwd")
	}

	return parsedUsers
}

// ParseUsers reads and parses users from the given passwd-format file.
func ParseUsers(path string) {
	parsedUsers = nil

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		u, err := parseLine(scanner.Text())
		if err != nil {
			_ = f.Close()
			log.Fatal(err)
		}

		parsedUsers = append(parsedUsers, u)
	}
	_ = f.Close()
}

func parseLine(line string) (User, error) {
	fs := strings.Split(line, ":")

	if len(fs) != 7 {
		return User{}, errors.New("unexpected number of fields in /etc/passwd")
	}

	// Parse the GECOS field
	gecos := strings.Split(fs[4], ",")

	u := User{}
	u.Details.Uid = fs[2]
	u.Details.Gid = fs[3]
	u.Details.Username = fs[0]
	u.Details.Name = gecos[0]
	u.Details.HomeDir = fs[5]
	u.Groups = parseGroups(u.Details)

	return u, nil
}

func parseGroups(currentUser user.User) []*user.Group {
	groupIDs, _ := currentUser.GroupIds()
	groups := make([]*user.Group, 0, len(groupIDs))

	for _, groupID := range groupIDs {
		foundGroup, _ := user.LookupGroupId(groupID)
		groups = append(groups, foundGroup)
	}

	return groups
}
