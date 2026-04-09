# ugm

<p>
    <a href="https://github.com/mrofisr/ugm-tui/releases"><img src="https://img.shields.io/github/v/release/mrofisr/ugm-tui" alt="Latest Release"></a>
    <a href="https://goreportcard.com/report/github.com/mrofisr/ugm-tui"><img src="https://goreportcard.com/badge/mrofisr/ugm-tui" alt="Go ReportCard"></a>
    <a href="https://github.com/mrofisr/ugm-tui/actions"><img src="https://github.com/mrofisr/ugm-tui/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

A terminal user interface (TUI) to view and manage UNIX users and groups.

## Features

- Browse system users and groups with fuzzy search
- View user details (UID, GID, home directory, group memberships)
- View group details and member lists
- **Account status badges** — ✅ active, 🔒 locked, ⚠️ expiring — visible in the user list
- **Last login** display in user detail view
- **System user filter** — press `s` to toggle system accounts (UID < 1000)
- **Create new users** with custom shell and password or SSH key authentication
- **Delete users** and their home directory (`userdel -r`)
- **Lock/unlock users** to revoke or restore access (`usermod --lock/--unlock`)
- **Set account expiry** for time-limited access (`chage --expiredate`)
- **Add/remove users from groups** for role-based access (`usermod -aG`, `gpasswd -d`)
- **Create/delete groups** to define roles (`groupadd`, `groupdel`)
- **View password aging info** (`chage -l`)
- **Command preview** — see the exact command before confirming destructive actions
- **Audit log** — all actions logged to `/var/log/ugm-audit.log` with timestamp and operator

## Requirements

- Linux, FreeBSD, OpenBSD, or NetBSD
- Go 1.25+ (to build from source)
- **Root privileges** — run with `sudo ugm`

## Installation

### From source

```sh
go install github.com/mrofisr/ugm-tui/cmd/ugm@latest
```

### From releases

Download a prebuilt binary from [releases](https://github.com/mrofisr/ugm-tui/releases).

## Usage

```sh
sudo ugm
```

### Navigation

| Key | Description |
|-----|-------------|
| `Ctrl+c` / `q` / `Esc` | Exit |
| `Tab` | Switch between user and group view |
| `↑` / `k` | Previous item |
| `↓` / `j` | Next item |
| `←` / `h` | Previous page |
| `→` / `l` | Next page |
| `/` | Search |
| `Enter` | Apply search |
| `m` | Manage selected user |
| `s` | Toggle system users (UID < 1000) |

### Management Actions

Press `m` on a selected user to open the management menu:

| Action | Description |
|--------|-------------|
| **Create New User** | Create a user with a custom shell. Authenticate via password or SSH public key. |
| **Delete User** | Remove user and their home directory (`userdel -r`). |
| **Lock User** | Disable login via `usermod --lock` (revoke access). |
| **Unlock User** | Re-enable login via `usermod --unlock`. |
| **Set Expiry Date** | Set account expiry via `chage --expiredate`. The OS locks the account automatically when expired. |
| **Add to Group** | Assign a role by adding user to a group (`usermod -aG`). |
| **Remove from Group** | Revoke a role by removing user from a group (`gpasswd -d`). |
| **View Password Aging** | Display password aging info via `chage -l`. |
| **Create Group** | Create a new group/role (`groupadd`). |
| **Delete Group** | Remove a group (`groupdel`). |

#### Management Navigation

| Key | Description |
|-----|-------------|
| `↑` / `k` | Previous menu item |
| `↓` / `j` | Next menu item |
| `Enter` | Select action / Submit form |
| `Tab` | Next field (in forms) / Toggle auth method |
| `y` / `n` | Confirm / Cancel (delete, lock, unlock) |
| `Esc` | Back to previous view |

### Audit Log

All management actions are logged to `/var/log/ugm-audit.log`:

```
2026-04-09T16:30:00+07:00 operator=abdur action=create-user target=rafi useradd -m -s /bin/bash rafi
2026-04-09T16:31:00+07:00 operator=abdur action=lock target=rafi usermod --lock rafi
```

The operator is detected from `$SUDO_USER`.

## Development

### Prerequisites

```sh
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Build

```sh
make build
```

### Format & Lint

```sh
make fmt     # format with gofumpt
make lint    # run golangci-lint
make fix     # auto-fix lint issues
make all     # fmt + fix + build
```

### Project Structure

```
.
├── cmd/ugm/main.go          # Entry point, root check, OS check
└── internal/
    ├── passwd/              # Parses /etc/passwd
    ├── group/               # Parses /etc/group
    ├── usermgmt/            # User management (create, delete, lock, unlock,
    │                        #   expiry, groups, SSH keys, password aging)
    ├── audit/               # Action audit logging
    └── tui/
        ├── tui.go           # Root model, state management
        ├── style.go         # Shared styles
        ├── userview.go      # User list view
        ├── groupview.go     # Group list view
        └── manageview.go    # Management actions
```

## Platform Notes

`ugm` only works on UNIX-based operating systems (Linux, FreeBSD, OpenBSD, NetBSD).

On macOS, the information reported will not be accurate. The tool relies on `/etc/passwd` and `/etc/group`, which are only consulted in single-user mode. macOS uses [Directory Services](https://developer.apple.com/documentation/devicemanagement/directoryservice) to manage users and groups.

## Built With

- [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles v2](https://github.com/charmbracelet/bubbles) — TUI components (list, viewport, table, text input)
- [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) — Terminal styling and layout

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Security

See [SECURITY.md](SECURITY.md) for the security policy.

## License

[MIT](LICENSE)
