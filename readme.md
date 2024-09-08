# Clash Rule Manager

A Go application to manage Clash proxy rules by updating GitHub Gists and Clash providers.

## Features

- Update rules for different categories (global-tv, direct, reject, trusted)
- Modify GitHub Gists with new domain rules
- Update Clash providers via API
- Check rule update status

## Usage

```
clash-rule <rule-type> <domain>
```

Supported rule types: global-tv, direct, reject, trusted

## Requirements

- Go 1.16+
- GitHub personal access token (PAT)
- Clash API access (configure URL and auth token in the code)

## Installation

1. Clone the repository
2. Set up your GitHub PAT:
   - Option 1: Set environment variable: `export GITHUB_TOKEN=your_token_here`
   - Option 2: Store in `~/.gist_pat` file
3. Build the application: `go build -o clash-rule`

## Configuration

Edit the constants in `main.go` to set:

- Gist ID
- Clash API URL
- Clash API auth token

## GitHub PAT

The application will look for the GitHub PAT in the following order:
1. `GITHUB_TOKEN` environment variable
2. `~/.gist_pat` file

If using the file method, ensure it contains only the PAT and has appropriate permissions.

## License

[MIT License](LICENSE)
