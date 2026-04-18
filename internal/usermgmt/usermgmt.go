//go:build linux || freebsd || openbsd || netbsd

// Package usermgmt provides functions for managing UNIX user accounts.
package usermgmt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CreateUser creates a new system user with the given shell.
func CreateUser(username, shell string) error {
	if shell == "" {
		shell = "/bin/bash"
	}
	return run("useradd", "-m", "-s", shell, username)
}

// DeleteUser removes a user and their home directory.
func DeleteUser(username string) error {
	return run("userdel", "-r", username)
}

// SetPassword sets the login password for the given user.
func SetPassword(username, password string) error {
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(username + ":" + password)
	return cmd.Run()
}

// SetSSHKey writes the given public key to the user's authorized_keys file.
func SetSSHKey(username, pubkey string) error {
	out, err := exec.Command("getent", "passwd", username).Output()
	if err != nil {
		return fmt.Errorf("user %s not found: %w", username, err)
	}

	fields := strings.Split(strings.TrimSpace(string(out)), ":")
	if len(fields) < 6 {
		return fmt.Errorf("unexpected passwd entry for %s", username)
	}

	homeDir := fields[5]
	uid, err := strconv.Atoi(fields[2])
	if err != nil {
		return fmt.Errorf("parse uid for %s: %w", username, err)
	}
	gid, err := strconv.Atoi(fields[3])
	if err != nil {
		return fmt.Errorf("parse gid for %s: %w", username, err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("create .ssh dir: %w", err)
	}

	authFile := filepath.Join(sshDir, "authorized_keys")
	if err := os.WriteFile(authFile, []byte(pubkey+"\n"), 0o600); err != nil {
		return fmt.Errorf("write authorized_keys: %w", err)
	}

	if err := os.Chown(sshDir, uid, gid); err != nil {
		return fmt.Errorf("chown .ssh: %w", err)
	}
	if err := os.Chown(authFile, uid, gid); err != nil {
		return fmt.Errorf("chown authorized_keys: %w", err)
	}
	return nil
}

// LockUser disables login for the given user via usermod --lock.
func LockUser(username string) error {
	return run("usermod", "--lock", username)
}

// UnlockUser re-enables login for the given user via usermod --unlock.
func UnlockUser(username string) error {
	return run("usermod", "--unlock", username)
}

// SetExpiry sets the account expiry date (YYYY-MM-DD) for the given user.
func SetExpiry(username, date string) error {
	return run("chage", "--expiredate", date, username)
}

// AddToGroup adds a user to a supplementary group.
func AddToGroup(username, groupname string) error {
	return run("usermod", "-aG", groupname, username)
}

// RemoveFromGroup removes a user from a group.
func RemoveFromGroup(username, groupname string) error {
	return run("gpasswd", "-d", username, groupname)
}

// CreateGroup creates a new system group.
func CreateGroup(name string) error {
	return run("groupadd", name)
}

// DeleteGroup removes a system group.
func DeleteGroup(name string) error {
	return run("groupdel", name)
}

// PasswordAging returns the output of chage -l for the given user.
func PasswordAging(username string) (string, error) {
	out, err := exec.Command("chage", "-l", username).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("chage -l %s: %s: %w", username, strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// LastLogin returns the last login time string for a user, or "Never" if none.
func LastLogin(username string) string {
	out, err := exec.Command("lastlog", "-u", username).CombinedOutput()
	if err != nil {
		return "Never"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "Never"
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 || fields[1] == "**Never" {
		return "Never"
	}
	return strings.Join(fields[3:], " ")
}

// AccountStatus returns a status string: "locked", "expires <date>", or "active".
func AccountStatus(username string) string {
	// Check locked via passwd -S: second field is "L" if locked
	if out, err := exec.Command("passwd", "-S", username).CombinedOutput(); err == nil {
		fields := strings.Fields(string(out))
		if len(fields) >= 2 && fields[1] == "L" {
			return "locked"
		}
	}

	// Check expired via chage -l: "Account expires" line
	out, err := exec.Command("chage", "-l", username).CombinedOutput()
	if err != nil {
		return "active"
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		val, found := strings.CutPrefix(line, "Account expires")
		if !found {
			continue
		}
		val, _ = strings.CutPrefix(val, ":")
		val = strings.TrimSpace(val)
		if val != "never" && val != "" {
			return "expires " + val
		}
	}

	return "active"
}

func run(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %s: %w", name, strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}
