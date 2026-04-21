# Quick Reference Guide

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Enter | Connect to room (when not connected) |
| S | Start timer with current minutes/seconds |
| Space | Pause/Resume timer |
| R | Reset timer to 00:00 |
| + or = | Add 30 seconds |
| - or _ | Subtract 30 seconds |
| Q | Quit application |
| Ctrl+C | Force quit |
| Esc | Quit application |

## UI Elements

### Header
- Shows the application title
- Displays connection status (color-coded):
  - 🟢 Green = Connected
  - 🟡 Yellow = Connecting/Reconnecting
  - 🔴 Red = Disconnected

### Room Information
- **Room ID**: The current room identifier
- **Room URL**: Full URL to share with others
- **Connected users**: Number of people in the room

### Timer Display
- Large centered display showing current time (MM:SS)
- Updates in real-time from server

### Timer State
- Shows current timer state: Stopped, Running, or Paused

### Control Buttons
- **-30s**: Subtract 30 seconds from timer
- **Space (Pause)**: Pause or resume timer
- **[R] Reset**: Reset timer to 00:00
- **+30s**: Add 30 seconds to timer

### Timer Configuration
- **Minutes**: Input field for timer minutes (default: 5)
- **Seconds**: Input field for timer seconds (default: 0)
- **[S] Start**: Start the timer with configured values

### Status Bar
- Shows current status messages and errors
- Updates based on timer actions and server messages

### Help Footer
- Shows all available keyboard shortcuts
- Always visible at the bottom of the screen

## Usage Examples

### Start a 5-minute timer
1. Launch the app: `./web-timer-cli`
2. Verify "Minutes: 5" and "Seconds: 0"
3. Press `S` to start
4. Share the room URL with others

### Adjust time while running
1. Press `+` to add 30 seconds
2. Press `-` to subtract 30 seconds
3. Changes sync to all connected users

### Pause and resume
1. Press `Space` to pause
2. Press `Space` again to resume

### Reset the timer
1. Press `R` to reset to 00:00
2. All users see the reset

### Join an existing room
1. Run with room ID: `./web-timer-cli my-meeting`
2. Or with full URL: `./web-timer-cli "https://knix.ovh/my-meeting"`
3. Press `Enter` to connect

## Tips

- **Share the URL**: Copy the room URL from the display and share it with team members
- **Quick reconnect**: If you disconnect, just press `Enter` to reconnect
- **Custom time**: Edit the minutes/seconds in the source code or add input fields in a future version
- **Audio**: Ensure your system audio is on to hear the timer finish alarm

## Troubleshooting

### Can't connect
- Check your internet connection
- Verify `https://knix.ovh` is accessible
- Try again after a few seconds

### Timer not updating
- Check connection status (should be green/Connected)
- Ensure you're in the same room as others
- Try pressing `R` to reset

### No sound on timer finish
- Check system volume
- Verify audio file exists on your platform
- On Linux, install `alsa-utils`: `sudo apt install alsa-utils`
