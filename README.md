# ugm

<p>
    <a href="https://github.com/ariasmn/ugm/releases"><img src="https://img.shields.io/github/v/release/ariasmn/ugm" alt="Latest Release"></a>
    <a href="https://goreportcard.com/report/github.com/ariasmn/ugm"><img src="https://goreportcard.com/badge/ariasmn/ugm" alt="Go ReportCard"></a>
    <a href="https://github.com/ariasmn/ugm/actions"><img src="https://github.com/ariasmn/ugm/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

A terminal user interface (TUI) to view and manage UNIX users and groups.

## Features

- Browse system users and groups with fuzzy search
- View user details (UID, GID, home directory, group memberships)
- View group details and member lists
- **Create new users** with custom shell and password or SSH key authentication
- **Lock users** to revoke access (`usermod --lock`)
- **Set account expiry** for time-limited access (`chage --expiredate`)

## Requirements

- Linux, FreeBSD, OpenBSD, or NetBSD
- Go 1.25+ (to build from source)
- **Root privileges** — run with `sudo ugm`

## Installation

### From source

```sh
go install github.com/ariasmn/ugm@latest
```

### From releases

Download a prebuilt binary from [releases](https://github.com/ariasmn/ugm/releases).

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

### Management Actions

Press `m` on a selected user to open the management menu:

| Action | Description |
|--------|-------------|
| **Create New User** | Create a user with a custom shell. Authenticate via password or SSH public key. |
| **Lock User** | Disable login via `usermod --lock` (revoke access). |
| **Set Expiry Date** | Set account expiry via `chage --expiredate`. The OS locks the account automatically when expired. |

#### Management Navigation

| Key | Description |
|-----|-------------|
| `↑` / `k` | Previous menu item |
| `↓` / `j` | Next menu item |
| `Enter` | Select action / Submit form |
| `Tab` | Next field (in forms) / Toggle auth method |
| `y` / `n` | Confirm / Cancel (lock user) |
| `Esc` | Back to previous view |

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
├── main.go                  # Entry point, root check, OS check
├── userparser/              # Parses /etc/passwd
├── groupparser/             # Parses /etc/group
├── usermgmt/                # User management (create, lock, expiry, SSH keys)
└── internal/tui/
    ├── tui.go               # Root TUI model, state management
    ├── common/              # Shared styles and utilities
    ├── user/                # User list view
    ├── group/               # Group list view
    └── manage/              # Management actions (create, lock, expiry)
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

## License

[MIT](LICENSE) © Ismael Arias
