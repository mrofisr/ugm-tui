# Security Policy

## Important

`ugm` runs as **root** and executes system commands that modify user accounts:

| Command | Used For |
|---------|----------|
| `useradd` | Creating users |
| `userdel` | Deleting users and home directories |
| `usermod` | Locking/unlocking users, group membership |
| `chpasswd` | Setting passwords |
| `chage` | Setting account expiry, viewing password aging |
| `gpasswd` | Removing users from groups |
| `groupadd` | Creating groups |
| `groupdel` | Deleting groups |
| `getent` | Looking up user info for SSH key setup |
| `passwd -S` | Checking account lock status |
| `lastlog` | Retrieving last login times |

It also writes to `~/.ssh/authorized_keys` when setting up SSH key authentication.

### Audit Log

All management actions are logged to `/var/log/ugm-audit.log` with:
- Timestamp (RFC 3339)
- Operator (detected from `$SUDO_USER`)
- Action performed
- Target user/group
- Exact command executed

Only run `ugm` on systems where you trust the binary. Always verify the checksum when downloading from releases.

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately by opening a [security advisory](https://github.com/mrofisr/ugm-tui/security/advisories/new) on GitHub.

Do **not** open a public issue for security vulnerabilities.

## Supported Versions

Only the latest release is supported with security updates.
