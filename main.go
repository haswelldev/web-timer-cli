package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Message types ─────────────────────────────────────────────────────────────

type tickMsg time.Time
type connectionStateMsg ConnectionState
type timeUpdateMsg int
type timerStateMsg TimerState
type userCountMsg int
type statusMsg string

// ── Color palette (Catppuccin Mocha) ──────────────────────────────────────────

var (
	colorBase     = lipgloss.Color("#1E1E2E")
	colorSurface0 = lipgloss.Color("#313244")
	colorSurface1 = lipgloss.Color("#45475A")
	colorOverlay0 = lipgloss.Color("#6C7086")
	colorText     = lipgloss.Color("#CDD6F4")
	colorSubtext  = lipgloss.Color("#BAC2DE")
	colorBlue     = lipgloss.Color("#89B4FA")
	colorGreen    = lipgloss.Color("#A6E3A1")
	colorYellow   = lipgloss.Color("#F9E2AF")
	colorRed      = lipgloss.Color("#F38BA8")
	colorMauve    = lipgloss.Color("#CBA6F7")
)

// ── Click zones ───────────────────────────────────────────────────────────────

type zone struct {
	x1, x2, y int
	action     string
}

var viewZones []zone

// ── Bubble Tea lifecycle ──────────────────────────────────────────────────────

func (m TimerModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		connectToRoomCmd(m),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			for _, z := range viewZones {
				if msg.Y == z.y && msg.X >= z.x1 && msg.X < z.x2 {
					switch z.action {
					case "focus-min":
						m.focusField = FocusMinutes
						return m, nil
					case "focus-sec":
						m.focusField = FocusSeconds
						return m, nil
					case "focus-alarm-min":
						m.focusField = FocusAlarmMins
						return m, nil
					case "focus-alarm-sec":
						m.focusField = FocusAlarmSecs
						return m, nil
					case "minus30":
						if m.connectionState == Connected {
							return m, adjustTimeCmd(m, -30)
						}
					case "pause":
						if m.connectionState == Connected {
							return m, togglePauseCmd(m)
						}
					case "reset":
						if m.connectionState == Connected {
							return m, resetTimerCmd(m)
						}
					case "plus30":
						if m.connectionState == Connected {
							return m, adjustTimeCmd(m, 30)
						}
					case "start":
						if m.connectionState == Connected {
							return m, startTimerCmd(m)
						}
					}
				}
			}
			// click anywhere else defocuses
			m.focusField = FocusNone
		}

	case tea.KeyMsg:
		// When an input field is focused, handle digits/backspace/tab/enter/esc first.
		if m.focusField != FocusNone {
			k := msg.String()
			switch {
			case k >= "0" && k <= "9":
				switch m.focusField {
				case FocusMinutes:
					if len(m.minutesInput) < 3 {
						m.minutesInput += k
					}
				case FocusSeconds:
					if len(m.secondsInput) < 2 {
						m.secondsInput += k
					}
				case FocusAlarmMins:
					if len(m.alarmMinsInput) < 3 {
						m.alarmMinsInput += k
					}
				case FocusAlarmSecs:
					if len(m.alarmSecsInput) < 2 {
						m.alarmSecsInput += k
					}
				}
				return m, nil
			case k == "backspace":
				switch m.focusField {
				case FocusMinutes:
					if len(m.minutesInput) > 0 {
						m.minutesInput = m.minutesInput[:len(m.minutesInput)-1]
					}
				case FocusSeconds:
					if len(m.secondsInput) > 0 {
						m.secondsInput = m.secondsInput[:len(m.secondsInput)-1]
					}
				case FocusAlarmMins:
					if len(m.alarmMinsInput) > 0 {
						m.alarmMinsInput = m.alarmMinsInput[:len(m.alarmMinsInput)-1]
					}
				case FocusAlarmSecs:
					if len(m.alarmSecsInput) > 0 {
						m.alarmSecsInput = m.alarmSecsInput[:len(m.alarmSecsInput)-1]
					}
				}
				return m, nil
			case k == "tab":
				switch m.focusField {
				case FocusMinutes:
					m.focusField = FocusSeconds
				case FocusSeconds:
					m.focusField = FocusAlarmMins
				case FocusAlarmMins:
					m.focusField = FocusAlarmSecs
				default:
					m.focusField = FocusNone
				}
				return m, nil
			case k == "enter":
				m.focusField = FocusNone
				if m.connectionState == Connected {
					return m, startTimerCmd(m)
				}
				return m, nil
			case k == "esc" || k == "ctrl+c":
				m.focusField = FocusNone
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit

		case "enter":
			if m.connectionState != Connected {
				return m, connectToRoomCmd(m)
			}

		case "tab":
			m.focusField = FocusMinutes
			return m, nil

		case "s":
			if m.connectionState == Connected {
				return m, startTimerCmd(m)
			}

		case " ":
			if m.connectionState == Connected {
				return m, togglePauseCmd(m)
			}

		case "r":
			if m.connectionState == Connected {
				return m, resetTimerCmd(m)
			}

		case "+", "=":
			if m.connectionState == Connected {
				return m, adjustTimeCmd(m, 30)
			}

		case "-", "_":
			if m.connectionState == Connected {
				return m, adjustTimeCmd(m, -30)
			}
		}

	case tickMsg:
		if cmd := m.CheckChannels(); cmd != nil {
			return m, tea.Sequence(tickCmd(), cmd)
		}
		return m, tickCmd()

	case connectionStateMsg:
		m.SetConnectionState(ConnectionState(msg))
		if ConnectionState(msg) == Connected {
			m.SetStatusMessage("Connected — press s to start timer")
		} else if ConnectionState(msg) == Connecting {
			m.SetStatusMessage("Connecting to room…")
		}
		return m, nil

	case timeUpdateMsg:
		m.SetTime(int(msg))
		m.alarmTriggered = false
		return m, nil

	case timerStateMsg:
		m.SetTimerState(TimerState(msg))
		return m, nil

	case userCountMsg:
		m.SetUserCount(int(msg))
		return m, nil

	case statusMsg:
		m.SetStatusMessage(string(msg))
		return m, nil

	case signalRConnectedMsg:
		m.signalRClient = msg.client
		m.SetupSignalRHandlers()
		return m, func() tea.Msg { return connectionStateMsg(Connected) }
	}

	return m, nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func connectToRoomCmd(model TimerModel) tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return connectionStateMsg(Connecting) },
		func() tea.Msg {
			client := NewSignalRClient(model.baseURL, model.roomID)
			if err := client.Connect(); err != nil {
				return statusMsg(fmt.Sprintf("Connection failed: %v", err))
			}
			return signalRConnectedMsg{client: client}
		},
	)
}

type signalRConnectedMsg struct{ client *SignalRClient }

func startTimerCmd(model TimerModel) tea.Cmd {
	return func() tea.Msg {
		if model.signalRClient == nil {
			return statusMsg("Not connected")
		}
		minutes, seconds := 0, 0
		fmt.Sscanf(model.minutesInput, "%d", &minutes)
		fmt.Sscanf(model.secondsInput, "%d", &seconds)
		if err := model.signalRClient.StartTimer(minutes, seconds); err != nil {
			return statusMsg(fmt.Sprintf("Start failed: %v", err))
		}
		return statusMsg("Timer started")
	}
}

func togglePauseCmd(model TimerModel) tea.Cmd {
	return func() tea.Msg {
		if model.signalRClient == nil {
			return statusMsg("Not connected")
		}
		if err := model.signalRClient.TogglePause(); err != nil {
			return statusMsg(fmt.Sprintf("Pause failed: %v", err))
		}
		return statusMsg("Timer toggled")
	}
}

func resetTimerCmd(model TimerModel) tea.Cmd {
	return func() tea.Msg {
		if model.signalRClient == nil {
			return statusMsg("Not connected")
		}
		if err := model.signalRClient.ResetTimer(); err != nil {
			return statusMsg(fmt.Sprintf("Reset failed: %v", err))
		}
		return statusMsg("Timer reset")
	}
}

func adjustTimeCmd(model TimerModel, delta int) tea.Cmd {
	return func() tea.Msg {
		if model.signalRClient == nil {
			return statusMsg("Not connected")
		}
		if err := model.signalRClient.AdjustTime(delta); err != nil {
			return statusMsg(fmt.Sprintf("Adjust failed: %v", err))
		}
		return statusMsg(fmt.Sprintf("Time adjusted by %+d seconds", delta))
	}
}

func playAlarmCmd() tea.Cmd {
	return func() tea.Msg {
		playSystemSound()
		return nil
	}
}

func playSystemSound() {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	case "linux":
		exec.Command("aplay", "/usr/share/sounds/alsa/Front_Center.wav").Run()
	case "windows":
		exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer 'C:\\Windows\\Media\\notify.wav').PlaySync()").Run()
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m TimerModel) View() string {
	w := m.width
	if w < 60 {
		w = 60
	}
	h := m.height
	if h < 20 {
		h = 20
	}

	// centerStr centers a (possibly ANSI-styled) string within width w.
	centerStr := func(s string) string {
		vis := lipgloss.Width(s)
		if vis >= w {
			return s
		}
		return strings.Repeat(" ", (w-vis)/2) + s
	}

	divider := lipgloss.NewStyle().Foreground(colorSurface1).Render(strings.Repeat("─", w))

	// Inline button: coloured background, dark text, 2-space horizontal padding.
	mkBtn := func(label string, fg, bg lipgloss.Color) string {
		return lipgloss.NewStyle().
			Foreground(fg).Background(bg).
			Padding(0, 2).Bold(true).
			Render(label)
	}

	var lines []string

	// ── Header bar ────────────────────────────────────────────────────────────
	dot := lipgloss.NewStyle().Foreground(colorOverlay0).Render("  ·  ")

	titleSpan := lipgloss.NewStyle().Foreground(colorMauve).Bold(true).Render("Web Timer")
	roomSpan := lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render(m.roomID)

	statusColor := colorRed
	statusLabel := "Disconnected"
	switch m.connectionState {
	case Connected:
		statusColor, statusLabel = colorGreen, "Connected"
	case Connecting:
		statusColor, statusLabel = colorYellow, "Connecting…"
	case Reconnecting:
		statusColor, statusLabel = colorYellow, "Reconnecting…"
	}
	connSpan := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render("●  " + statusLabel)
	usersSpan := lipgloss.NewStyle().Foreground(colorOverlay0).Render(fmt.Sprintf("%d users", m.userCount))

	headerContent := titleSpan + dot + roomSpan + dot + connSpan + dot + usersSpan
	header := lipgloss.NewStyle().
		Background(colorSurface0).Width(w).Padding(0, 2).
		Render(headerContent)

	lines = append(lines, header) // row 0
	lines = append(lines, "")     // row 1

	// ── Timer box ─────────────────────────────────────────────────────────────
	timerFg, timerBorder := colorText, colorSurface1
	switch m.timerState {
	case Running:
		timerFg, timerBorder = colorGreen, colorGreen
	case Paused:
		timerFg, timerBorder = colorYellow, colorYellow
	}

	timerBox := lipgloss.NewStyle().
		Foreground(timerFg).Bold(true).
		Padding(1, 6).
		Border(lipgloss.RoundedBorder()).BorderForeground(timerBorder).
		Render(m.FormatTimer())

	for _, l := range strings.Split(timerBox, "\n") {
		lines = append(lines, centerStr(l))
	}
	// rows 2-6 (5 lines: top border, blank, time, blank, bottom border)

	lines = append(lines, "") // row 7

	// ── State indicator ───────────────────────────────────────────────────────
	stateStr := lipgloss.NewStyle().Foreground(colorOverlay0).Render("■  Stopped")
	switch m.timerState {
	case Running:
		stateStr = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("▶  Running")
	case Paused:
		stateStr = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("⏸  Paused")
	}
	lines = append(lines, centerStr(stateStr)) // row 8
	lines = append(lines, "")                  // row 9
	lines = append(lines, divider)             // row 10
	lines = append(lines, "")                  // row 11

	// ── Control buttons ───────────────────────────────────────────────────────
	btnRow := len(lines) // row 12

	btn1 := mkBtn("-30s", colorText, colorSurface1)
	btn2 := mkBtn("Pause", colorBase, colorYellow)
	btn3 := mkBtn("Reset", colorBase, colorRed)
	btn4 := mkBtn("+30s", colorText, colorSurface1)

	gap3 := "   "
	buttonRowStr := btn1 + gap3 + btn2 + gap3 + btn3 + gap3 + btn4
	brsVis := lipgloss.Width(buttonRowStr)
	brsMargin := (w - brsVis) / 2
	if brsMargin < 0 {
		brsMargin = 0
	}

	newZones := []zone{}
	bx := brsMargin
	for _, b := range []struct {
		btn    string
		action string
	}{
		{btn1, "minus30"}, {btn2, "pause"}, {btn3, "reset"}, {btn4, "plus30"},
	} {
		w1 := lipgloss.Width(b.btn)
		newZones = append(newZones, zone{bx, bx + w1, btnRow, b.action})
		bx += w1 + len(gap3)
	}

	lines = append(lines, strings.Repeat(" ", brsMargin)+buttonRowStr) // row 12
	lines = append(lines, "")                                           // row 13

	// ── Start row ─────────────────────────────────────────────────────────────
	startRow := len(lines) // row 14

	fieldStyle := func(val string, focused bool) string {
		s := lipgloss.NewStyle().Padding(0, 1)
		if focused {
			s = s.Foreground(colorBase).Background(colorBlue).Bold(true)
		} else {
			s = s.Foreground(colorText).Background(colorSurface1)
		}
		displayVal := val
		if displayVal == "" && !focused {
			displayVal = "0"
		}
		cursor := ""
		if focused {
			cursor = "▌"
		}
		return s.Render(displayVal + cursor)
	}

	minLabel := lipgloss.NewStyle().Foreground(colorOverlay0).Render("Min")
	minVal := fieldStyle(m.minutesInput, m.focusField == FocusMinutes)
	secLabel := lipgloss.NewStyle().Foreground(colorOverlay0).Render("Sec")
	secVal := fieldStyle(m.secondsInput, m.focusField == FocusSeconds)
	btnStart := mkBtn("▶  Start", colorBase, colorGreen)

	const startIndent = 4
	minLabelX := startIndent
	minValX := minLabelX + lipgloss.Width(minLabel) + 1
	secLabelX := minValX + lipgloss.Width(minVal) + 4
	secValX := secLabelX + lipgloss.Width(secLabel) + 1

	newZones = append(newZones, zone{minValX, secLabelX - 1, startRow, "focus-min"})
	newZones = append(newZones, zone{secValX, secValX + lipgloss.Width(secVal) + 1, startRow, "focus-sec"})

	inputPart := minLabel + " " + minVal + "    " + secLabel + " " + secVal
	inputVis := lipgloss.Width(inputPart)
	startBtnVis := lipgloss.Width(btnStart)

	gap := w - startIndent - inputVis - startBtnVis - startIndent
	if gap < 4 {
		gap = 4
	}

	startBtnX := startIndent + inputVis + gap
	newZones = append(newZones, zone{startBtnX, startBtnX + startBtnVis, startRow, "start"})

	startLine := strings.Repeat(" ", startIndent) + inputPart + strings.Repeat(" ", gap) + btnStart
	lines = append(lines, startLine) // row 14
	lines = append(lines, "")        // row 15

	// ── Personal alarm row ────────────────────────────────────────────────────
	alarmRow := len(lines) // row 16

	bellIcon := lipgloss.NewStyle().Foreground(colorMauve).Render("🔔")
	alarmLabel := lipgloss.NewStyle().Foreground(colorOverlay0).Render("Personal alarm at")
	alarmMinLabel := lipgloss.NewStyle().Foreground(colorOverlay0).Render("Min")
	alarmMinVal := fieldStyle(m.alarmMinsInput, m.focusField == FocusAlarmMins)
	alarmSecLabel := lipgloss.NewStyle().Foreground(colorOverlay0).Render("Sec")
	alarmSecVal := fieldStyle(m.alarmSecsInput, m.focusField == FocusAlarmSecs)

	alarmMinValX := startIndent + lipgloss.Width(bellIcon) + 1 + lipgloss.Width(alarmLabel) + 1 + lipgloss.Width(alarmMinLabel) + 1
	alarmSecValX := alarmMinValX + lipgloss.Width(alarmMinVal) + 4 + lipgloss.Width(alarmSecLabel) + 1

	newZones = append(newZones, zone{alarmMinValX, alarmMinValX + lipgloss.Width(alarmMinVal), alarmRow, "focus-alarm-min"})
	newZones = append(newZones, zone{alarmSecValX, alarmSecValX + lipgloss.Width(alarmSecVal), alarmRow, "focus-alarm-sec"})

	alarmLine := strings.Repeat(" ", startIndent) + bellIcon + " " + alarmLabel + " " + alarmMinLabel + " " + alarmMinVal + "    " + alarmSecLabel + " " + alarmSecVal
	lines = append(lines, alarmLine) // row 16
	lines = append(lines, "")        // row 17
	lines = append(lines, divider)   // row 18
	lines = append(lines, "")        // row 19

	// ── Status message ────────────────────────────────────────────────────────
	icon := lipgloss.NewStyle().Foreground(colorBlue).Render("ℹ")
	statusText := lipgloss.NewStyle().Foreground(colorSubtext).Render(m.statusMessage)
	lines = append(lines, "  "+icon+"  "+statusText) // row 18

	// ── Fill to bottom ────────────────────────────────────────────────────────
	for len(lines) < h-2 {
		lines = append(lines, "")
	}

	// ── Help bar ──────────────────────────────────────────────────────────────
	lines = append(lines, divider)

	key := func(k string) string {
		return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render(k)
	}
	desc := func(d string) string {
		return lipgloss.NewStyle().Foreground(colorOverlay0).Render(d)
	}
	helpSep := lipgloss.NewStyle().Foreground(colorSurface1).Render("  ·  ")
	helpLine := strings.Join([]string{
		key("q") + desc(" quit"),
		key("tab") + desc(" edit min/sec"),
		key("s") + desc(" start"),
		key("space") + desc(" pause"),
		key("r") + desc(" reset"),
		key("+/-") + desc(" ±30s"),
	}, helpSep)
	lines = append(lines, "  "+helpLine)

	viewZones = newZones
	return strings.Join(lines, "\n")
}

// ── URL parsing + main ────────────────────────────────────────────────────────

func parseArg(arg string) (baseURL, roomID string) {
	if u, err := url.Parse(arg); err == nil && u.Scheme != "" && u.Host != "" {
		return u.Scheme + "://" + u.Host, strings.TrimPrefix(u.Path, "/")
	}
	return "https://knix.ovh", arg
}

func main() {
	baseURL := "https://knix.ovh"
	roomID := ""
	if len(os.Args) > 1 {
		baseURL, roomID = parseArg(os.Args[1])
	}

	model := NewTimerModel(baseURL, roomID)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
