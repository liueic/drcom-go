# Dr.COM CLI Tool (Go Edition)

A cross-platform command-line tool for managing Dr.COM laboratory network login, logout, and monitoring.

## Features

- **Single Binary**: No dependencies, easy to deploy.
- **Interactive Login**: Prompts for credentials if not configured.
- **Daemon Mode**: Auto-reconnects when the network drops.
- **Status Monitoring**: Shows traffic usage (with warnings) and balance.
- **Configurable**: Uses `~/.config/drcom-go/config.yaml`.

## Installation

```bash
# Clone and build
git clone https://github.com/liueic/drcom-go.git
cd drcom-go
go build -o drcom
sudo mv drcom /usr/local/bin/
```

## Usage

### 1. Login
```bash
drcom login
```
If it's your first time, it will ask for Host (default `http://10.10.10.9`), Username, and Password. It saves them to `~/.config/drcom-go/config.yaml`.

### 2. Check Status
```bash
drcom status
```
Shows current traffic and balance.

### 3. Logout
```bash
drcom logout
```

### 4. Daemon (Keep Alive)
Run this in the background (e.g., using `nohup` or `systemd`) to ensure you stay online.
```bash
drcom daemon
```

## Configuration
File: `~/.config/drcom-go/config.yaml`

```yaml
auth:
  host: http://10.10.10.9
  username: "123456"
  password: "password"
daemon:
  interval: 60
```
