# Copilot Coding Agent Instructions

This document provides instructions for AI coding agents working on this repository.

## Project Overview

TODO

## Technology Stack

- **Go**: 1.21+ required
- **Validation**: [go-playground/validator/v10](https://github.com/go-playground/validator)

## Directory Structure

- cmd: Entry points for various subcommands in the CLI.
- models: Structs for domain entities used across the application.
- pkg: Shared packages that may be used by other applications.
- vendor: Application dependencies.

```
./
├── cmd
│   ├── root.go
│   ├── subcommand1.go
│   └── subcommand2.go
├── internal
│   ├── fileutils
│   └── apiutils
├── models
├── pkg
└── vendor
```

## Development Commands

```bash
# Install dependencies
go mod download

# Build
go build ./...

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Start the server
go run main.go
```

## Code Style Guidelines

1. **Error Handling**: Always handle errors explicitly. Do not ignore errors with `_`.

## Testing

- Tests are in `*_test.go` files alongside source code
- Run `go test ./...` before committing changes


## Security Considerations

- Always validate user input using validator tags