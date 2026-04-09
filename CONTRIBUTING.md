# Contributing to ugm

Thanks for your interest in contributing! Here's how to get started.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```sh
   git clone https://github.com/<your-username>/ugm-tui.git
   cd ugm-tui
   ```
3. Install development tools:
   ```sh
   go install mvdan.cc/gofumpt@latest
   go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
   ```

## Development Workflow

### Build

```sh
make build
```

The binary is built from `cmd/ugm/main.go` and output as `./ugm`.

### Format & Lint

All code must pass formatting and linting before submission:

```sh
make fmt     # format with gofumpt
make lint    # run golangci-lint
make fix     # auto-fix lint issues
make all     # fmt + fix + build
```

`make build` depends on `make lint` — the build will fail if linting doesn't pass.

### Run Tests

```sh
go test ./...
```

Note: some tests require root privileges since they interact with `/etc/passwd` and `/etc/group`.

## Code Style

This project follows the [Google Go Style Guide](https://google.github.io/styleguide/go/):

- Entry point in `cmd/ugm/`; all packages under `internal/`
- Return errors from functions — no `log.Fatal` in library code
- No global mutable state — data is passed, not cached in package vars
- Unexported package-level vars/consts prefixed with `_` (e.g. `_defaultShell`)
- Import grouping: stdlib → external → internal (separated by blank lines)
- Doc comments on all exported types and functions

## Submitting Changes

1. Create a feature branch from `main`:
   ```sh
   git checkout -b feature/my-change
   ```
2. Make your changes
3. Run `make all` to ensure formatting, linting, and build pass
4. Run `go test ./...` to ensure tests pass
5. Commit with a clear message:
   ```
   feat: add user deletion support
   fix: handle empty group members
   docs: update keybinding table
   ```
6. Push and open a pull request

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR
- Add doc comments to all exported types and functions
- Ensure `make lint` reports 0 issues
- Update `README.md` if you add new features or keybindings
- All management actions must write to the audit log
- Don't break existing tests

## Project Structure

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

## Reporting Bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- OS and Go version (`go version`)

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.
