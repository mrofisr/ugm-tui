# Contributing to ugm

Thanks for your interest in contributing! Here's how to get started.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```sh
   git clone https://github.com/<your-username>/ugm.git
   cd ugm
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

### Format & Lint

All code must pass formatting and linting before submission:

```sh
make fmt     # format with gofumpt
make lint    # run golangci-lint
make fix     # auto-fix lint issues
make all     # fmt + fix + build
```

### Run Tests

```sh
go test ./...
```

Note: some tests require root privileges since they interact with `/etc/passwd` and `/etc/group`.

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
- Don't break existing tests

## Project Structure

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

## Reporting Bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- OS and Go version (`go version`)

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.
