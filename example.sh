#!/bin/bash

# Example script demonstrating web-timer-cli usage

echo "Web Timer CLI - Example Usage"
echo "==============================="
echo ""

# Check if the binary exists
if [ ! -f "./web-timer-cli" ]; then
    echo "Error: web-timer-cli binary not found."
    echo "Please build it first with: go build -o web-timer-cli"
    exit 1
fi

echo "This script demonstrates various ways to use the web-timer-cli."
echo ""
echo "1. Join a random room (this will start the app):"
echo "   ./web-timer-cli"
echo ""
echo "2. Join a specific room:"
echo "   ./web-timer-cli my-meeting"
echo ""
echo "3. Join using a full URL:"
echo "   ./web-timer-cli 'https://knix.ovh/standup-2024'"
echo ""
echo "4. Common workflows:"
echo "   a) Start a 5-minute timer:"
echo "      - Launch app with: ./web-timer-cli daily-standup"
echo "      - Press Enter to connect"
echo "      - Press S to start (default 5:00)"
echo "      - Share the displayed URL with team"
echo ""
echo "   b) Pomodoro timer (25 minutes):"
echo "      - Launch app: ./web-timer-cli pomodoro"
echo "      - Press Enter to connect"
echo "      - Edit source to change default to 25 minutes, or use +/- to adjust"
echo "      - Press S to start"
echo ""
echo "   c) Quick sync with existing session:"
echo "      - Someone shares URL: https://knix.ovh/abc123"
echo "      - Join with: ./web-timer-cli abc123"
echo "      - Press Enter to connect"
echo "      - You'll see the same timer as everyone else"
echo ""
echo "Keyboard Controls:"
echo "  Enter  - Connect to room"
echo "  S      - Start timer"
echo "  Space  - Pause/Resume"
echo "  R      - Reset timer"
echo "  +      - Add 30 seconds"
echo "  -      - Subtract 30 seconds"
echo "  Q      - Quit"
echo ""
echo "Press Ctrl+C to quit this script."
echo ""
echo "Would you like to launch the app now? (y/n)"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "Launching web-timer-cli..."
    ./web-timer-cli
else
    echo "You can launch it later with: ./web-timer-cli"
fi
