# Web Timer CLI

[![Build](https://github.com/haswelldev/web-timer-cli/actions/workflows/release.yml/badge.svg)](https://github.com/haswelldev/web-timer-cli/actions/workflows/release.yml)

A cross-platform terminal client for the [Web Timer](https://knix.ovh) shared countdown app. Multiple people can share and control the same timer in real time — from any browser or terminal.

## Install

### macOS / Linux (curl)

```sh
curl -fsSL https://raw.githubusercontent.com/haswelldev/web-timer-cli/main/install.sh | sh
```

### macOS / Linux (wget)

```sh
wget -qO- https://raw.githubusercontent.com/haswelldev/web-timer-cli/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
iex (Invoke-WebRequest https://raw.githubusercontent.com/haswelldev/web-timer-cli/main/install.ps1).Content
```

### From source

```sh
go install github.com/athened/web-timer-cli@latest
```

### Manual download

Download the binary for your platform from the [latest release](https://github.com/haswelldev/web-timer-cli/releases/latest):

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `web-timer-cli-darwin-arm64` |
| macOS (Intel) | `web-timer-cli-darwin-amd64` |
| Linux (x86_64) | `web-timer-cli-linux-amd64` |
| Linux (ARM64) | `web-timer-cli-linux-arm64` |
| Windows (x86_64) | `web-timer-cli-windows-amd64.exe` |
| Windows (ARM64) | `web-timer-cli-windows-arm64.exe` |

## Usage

```sh
# Join a random room on knix.ovh
web-timer-cli

# Join a specific room by ID
web-timer-cli my-room

# Join by full URL (extracts server and room automatically)
web-timer-cli https://knix.ovh/my-room
```

Share the room URL displayed in the header with others — they can join from the web app or another terminal.

## Controls

### Keyboard

| Key | Action |
|-----|--------|
| `Tab` | Focus timer input (Min → Sec → Alarm Min → Alarm Sec) |
| `S` | Start timer with entered Min/Sec |
| `Enter` | Start timer (when input focused) |
| `Space` | Pause / Resume |
| `R` | Reset timer |
| `+` / `=` | Add 30 seconds |
| `-` / `_` | Subtract 30 seconds |
| `Q` / `Esc` / `Ctrl+C` | Quit |

### Mouse

Click any button or input field directly. Input fields become editable on click; type digits and use Backspace.

### Personal Alarm

Set a personal alarm threshold (default 0 min 5 sec). When the shared timer ticks down to that value your system sound plays — useful when you're not watching the screen.

## Uninstall

```sh
sudo rm /usr/local/bin/web-timer-cli
```

Windows: delete `web-timer-cli.exe` from where you placed it.

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket client

## Credits

Based on [Web Timer](https://github.com/ChocoWhoopies/web-timer) by ChocoWhoopies.
