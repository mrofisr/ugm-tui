# Security Policy

## Important

`ugm` runs as **root** and executes system commands that modify user accounts:

| Command | Used For |
|---------|----------|
| `useradd` | Creating users |
| `userdel` | Deleting users |
| `usermod` | Locking/unlocking users, group membership |
| `chpasswd` | Setting passwords |
| `chage` | Setting account expiry, viewing password aging |
| `gpasswd` | Removing users from groups |
| `getent` | Looking up user info for SSH key setup |

It also writes to `~/.ssh/authorized_keys` when setting up SSH key authentication.

Only run `ugm` on systems where you trust the binary. Always verify the checksum when downloading from releases.

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately by opening a [security advisory](https://github.com/mrofisr/ugm-tui/security/advisories/new) on GitHub.

Do **not** open a public issue for security vulnerabilities.

## Supported Versions

Only the latest release is supported with security updates.
