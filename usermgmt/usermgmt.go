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

// SetPassword sets the login password for the given user.
func SetPassword(username, password string) error {
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	return cmd.Run()
}

// SetSSHKey writes the given public key to the user's authorized_keys file.
func SetSSHKey(username, pubkey string) error {
	u, err := exec.Command("getent", "passwd", username).Output()
	if err != nil {
		return fmt.Errorf("user %s not found: %w", username, err)
	}
	fields := strings.Split(strings.TrimSpace(string(u)), ":")
	if len(fields) < 6 {
		return fmt.Errorf("unexpected passwd entry for %s", username)
	}
	homeDir := fields[5]
	uid, _ := strconv.Atoi(fields[2])
	gid, _ := strconv.Atoi(fields[3])

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return err
	}
	authFile := filepath.Join(sshDir, "authorized_keys")
	if err := os.WriteFile(authFile, []byte(pubkey+"\n"), 0o600); err != nil {
		return err
	}
	if err := os.Chown(sshDir, uid, gid); err != nil {
		return err
	}
	return os.Chown(authFile, uid, gid)
}

// LockUser disables login for the given user via usermod --lock.
func LockUser(username string) error {
	return run("usermod", "--lock", username)
}

// SetExpiry sets the account expiry date (YYYY-MM-DD) for the given user.
func SetExpiry(username, date string) error {
	return run("chage", "--expiredate", date, username)
}

func run(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
