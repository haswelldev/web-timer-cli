# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `make build` — build the `web-timer-cli` binary for the current platform (injects `main.Version` via `-ldflags`; `web-timer-cli --version` prints it).
- `make run` — build and run.
- `make install` / `make update` / `make uninstall` — build and install (or remove) the binary system-wide under `$(PREFIX)/bin` (default `/usr/local/bin`; `sudo` only if not writable). `update` rebuilds the current tree and overwrites the installed binary. Override with `PREFIX=…`.
- `make test` or `go test -v ./...` — run the full test suite.
- `go test -v -run TestTimerModelFormatTimer` — run a single test by name.
- `make fmt` — `go fmt ./...`.
- `make lint` — runs `golangci-lint run` (not vendored; must be installed separately).
- `make deps` — `go mod download && go mod tidy`.
- `make build-all` — cross-compile for linux/amd64, darwin/amd64, darwin/arm64, windows/amd64.

## Architecture

Single-package Go program (`package main`) split across four files. The Bubble Tea Elm-style loop in `main.go` is the only place that mutates UI state; the SignalR client runs on its own goroutine and communicates upward through buffered channels.

### Files and their roles

- **main.go** — Bubble Tea `Init`/`Update`/`View`, keybindings, command constructors (`connectToRoomCmd`, `startTimerCmd`, etc.), and `playSystemSound` (OS-specific: `afplay` on darwin, `aplay` on linux, PowerShell `SoundPlayer` on windows).
- **model.go** — `TimerModel` struct, state enums (`TimerState`, `ConnectionState`), and the channel-bridge between SignalR callbacks and the Bubble Tea loop. `SetupSignalRHandlers` registers handlers that push into `timeChan`/`userCountChan`/`messageChan`; `CheckChannels` is called every tick from `Update` to drain them without blocking. The server sends no timer-state event, so `CheckChannels` **infers** Running/Stopped/Paused from the stream of time updates (counting down ⇒ Running; zero ⇒ Stopped; stream stalled >2s while >0 ⇒ Paused), using `lastTimeUpdate`.
- **signalr.go** — hand-rolled SignalR JSON-protocol client over `gorilla/websocket`. Performs the protocol-1 handshake, frames messages with the `0x1E` record separator, and dispatches inbound invocations (`type: 1`) to registered handlers. A `keepAlive` goroutine sends a type-6 ping every 15s so the server doesn't drop the socket; all writes go through `writeFrame` (serialized by `writeMutex`, since gorilla/websocket forbids concurrent writers). The `handlers` map is guarded by `handlersMutex`. Message types 2–6 are ignored; type 7 (close) tears down the connection.
- **model_test.go** — unit tests for timer formatting, state transitions, room ID generation, and client construction. No network tests.

### Key flow

1. `main` parses an optional arg via `parseArg` — full URL extracts `baseURL`+roomID, bare string is a roomID against `https://knix.ovh`. Constructs `NewTimerModel` (random `room-XXXXXXXX` if none given).
2. `Init()` immediately calls `connectToRoomCmd` (no Enter needed). `Update` receives `signalRConnectedMsg`, stores the client, calls `SetupSignalRHandlers`.
3. `SignalRClient.Connect` POSTs to `/timerHub/negotiate?negotiateVersion=1`, upgrades `https`→`wss` with the token, sends `{"protocol":"json","version":1}\x1E` handshake, then invokes `JoinSession(roomID)` and spawns `readMessages`.
4. Server events → `readMessages` → `processMessage` (splits on `0x1E`) → `handleInvocation` → registered handler → channel send. The 1-second tick drains channels via `CheckChannels` on the UI goroutine.
5. Two alarms: `alarmTriggered` fires when the shared timer reaches zero; `personalAlarmFired` fires when the timer ticks to or below the user-configured threshold (`alarmMinsInput`/`alarmSecsInput`, default 0m5s).
6. On each tick, `Update` checks for a dropped socket (`connectionState == Connected` but `!client.IsConnected()`) and for a pending retry (`shouldReconnect`/`reconnectCountdown`), re-running `connectToRoomCmd`. A failed connect returns `connectFailedMsg`, which sets `Disconnected` and schedules a retry (rather than leaving the UI stuck on "Connecting…").

### Input fields and focus

`FocusField` enum: `FocusNone → FocusMinutes → FocusSeconds → FocusAlarmMins → FocusAlarmSecs`. Tab cycles through this order; clicking a field sets focus directly. Fields allow empty string while focused (displayed as cursor-only); unfocused empty fields render as "0". `startTimerCmd` defaults to `0` for any field that parses as 0 or is empty. Pressing Enter starts the timer **only** when a Min/Sec field is focused; when a personal-alarm field is focused Enter just confirms the value and defocuses (it must not start/restart the shared timer). Start/pause/reset also apply an optimistic local `timerState` (`optimisticStart`/`optimisticReset`/`optimisticTogglePause`) for instant feedback before the server echo arrives.

### Conventions worth knowing

- All SignalR outbound invocations carry `roomID` as the first arg (see `StartTimer`, `TogglePause`, `ResetTimer`, `AdjustTime`). Add new methods the same way.
- Never send on the channels without a `default:` case — they're buffered size 10 and handlers must not block the reader goroutine.
- `connMutex` guards `conn`/`connected`; always snapshot `conn` under the lock before I/O.
- Record separator (`0x1E`) must be appended to every outbound frame and stripped from every inbound frame.
- `viewZones` is a package-level `[]zone` rebuilt every `View()` call; mouse click coordinates are resolved against it in `Update`.

### CI / Release

`.github/workflows/release.yml` builds 6 binaries (linux/darwin/windows × amd64/arm64) on every push to `main` and publishes them as the `latest` pre-release tag (overwriting the previous one).
