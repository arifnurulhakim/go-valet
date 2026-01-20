# go-valet

A Laravel Valet clone for Go.

## Prerequisites

- **Go** (1.20+)
- **Air** (Hot reload): `go install github.com/air-verse/air@latest`
- **Caddy** (Reverse Proxy): `brew install caddy` (or similar)
- **Dnsmasq** (for *.test domain): Configure `*.test` to point to `127.0.0.1`.

## Installation

```bash
git clone https://github.com/hubton/go-valet.git
cd go-valet
go install ./cmd/go-valet
```

## Architecture

- **Registry**: Stores configuration in `~/.config/go-valet/registry.json`.
- **Daemon**: Long-running process that:
  - Watches registry and parked directories.
  - Spawns apps using `air` on deterministic ports.
  - Generates a Caddyfile.
  - Reloads Caddy.
- **Port Allocation**: `9000 + hash(path) % 1000`. Deterministic.
- **Reverse Proxy**: Caddy proxies `app.test` -> `localhost:PORT`.

## Usage

### 1. Start the Daemon

```bash
# Foreground
go-valet daemon start

# Background (macOS launchd)
cp misc/launchd.plist ~/Library/LaunchAgents/com.github.hubton.go-valet.plist
launchctl load ~/Library/LaunchAgents/com.github.hubton.go-valet.plist
```

### 2. Park a Directory

Navigate to your workspace containing Go projects:

```bash
cd ~/code
go-valet park
```

Any folder inside `~/code` with a `main.go` (or `cmd/*/main.go`) will be served at `http://folder-name.test`.

### 3. Link a Single App

Navigate to a specific project:

```bash
cd ~/code/my-project
go-valet link my-app
```

Served at `http://my-app.test`.

### 4. Status

Check registered parks, links, and running apps:

```bash
go-valet status
```

### 5. Unlink/Unpark

```bash
go-valet unlink my-app
go-valet unpark
```

## Logs

Daemon logs: depends on how you run it (launchd: `/tmp/go-valet.*.log`).
App logs: `~/.config/go-valet/logs/<app-name>.log`.
