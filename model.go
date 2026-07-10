package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TimerState int

const (
	Stopped TimerState = iota
	Running
	Paused
)

type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
	Reconnecting
)

type FocusField int

const (
	FocusNone      FocusField = iota
	FocusMinutes              // typing into the Min field
	FocusSeconds              // typing into the Sec field
	FocusAlarmMins            // typing into the personal alarm Min field
	FocusAlarmSecs            // typing into the personal alarm Sec field
)

type TimerModel struct {
	roomID             string
	baseURL            string
	totalSeconds       int
	timerState         TimerState
	connectionState    ConnectionState
	userCount          int
	minutesInput       string
	secondsInput       string
	focusField         FocusField
	statusMessage      string
	alarmTriggered     bool
	lastTimeUpdate     time.Time
	alarmMinsInput     string
	alarmSecsInput     string
	personalAlarmFired bool
	shouldReconnect    bool
	reconnectCountdown int
	signalRClient      *SignalRClient
	timeChan           chan int
	userCountChan      chan int
	messageChan        chan string
	width              int
	height             int
}

func NewTimerModel(baseURL, roomID string) *TimerModel {
	if roomID == "" {
		roomID = generateRoomID()
	}
	return &TimerModel{
		roomID:          roomID,
		baseURL:         baseURL,
		totalSeconds:    0,
		timerState:      Stopped,
		connectionState: Disconnected,
		userCount:       0,
		minutesInput:    "5",
		secondsInput:    "0",
		alarmMinsInput:  "0",
		alarmSecsInput:  "5",
		statusMessage:   "Connecting to room…",
		alarmTriggered:  false,
		timeChan:        make(chan int, 10),
		userCountChan:   make(chan int, 10),
		messageChan:     make(chan string, 10),
		width:           80,
		height:          24,
	}
}

func (m *TimerModel) FormatTimer() string {
	minutes := m.totalSeconds / 60
	seconds := m.totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (m *TimerModel) ConnectionStateText() string {
	switch m.connectionState {
	case Connected:
		return "Connected"
	case Connecting:
		return "Connecting..."
	case Reconnecting:
		return "Reconnecting..."
	default:
		return "Disconnected"
	}
}

func (m *TimerModel) TimerStateText() string {
	switch m.timerState {
	case Running:
		return "Running"
	case Paused:
		return "Paused"
	default:
		return "Stopped"
	}
}

func (m *TimerModel) SetTime(totalSeconds int) {
	m.totalSeconds = totalSeconds
}

func (m *TimerModel) SetTimerState(state TimerState) {
	m.timerState = state
}

func (m *TimerModel) SetConnectionState(state ConnectionState) {
	m.connectionState = state
}

func (m *TimerModel) SetUserCount(count int) {
	m.userCount = count
}

func (m *TimerModel) SetStatusMessage(msg string) {
	m.statusMessage = msg
}

func (m *TimerModel) SetupSignalRHandlers() {
	if m.signalRClient == nil {
		return
	}

	m.signalRClient.RegisterHandler("ReceiveTime", func(args []interface{}) {
		if len(args) > 0 {
			if seconds, ok := args[0].(float64); ok {
				select {
				case m.timeChan <- int(seconds):
				default:
				}
			}
		}
	})

	m.signalRClient.RegisterHandler("TimerFinished", func(args []interface{}) {
		select {
		case m.timeChan <- 0:
		default:
		}
	})

	m.signalRClient.RegisterHandler("RoomUserCount", func(args []interface{}) {
		if len(args) > 0 {
			if count, ok := args[0].(float64); ok {
				select {
				case m.userCountChan <- int(count):
				default:
				}
			}
		}
	})

	m.signalRClient.RegisterHandler("Message", func(args []interface{}) {
		if len(args) > 0 {
			if msg, ok := args[0].(string); ok {
				select {
				case m.messageChan <- msg:
				default:
				}
			}
		}
	})
}

func (m *TimerModel) CheckChannels() tea.Cmd {
	var cmds []tea.Cmd

	select {
	case seconds := <-m.timeChan:
		prevSeconds := m.totalSeconds
		wasNonZero := prevSeconds > 0
		m.SetTime(seconds)

		// Infer timer state from the stream of time updates (the server sends
		// no explicit state event): counting down means Running, zero means
		// Stopped, a stalled stream is treated as Paused below.
		switch {
		case seconds == 0:
			m.timerState = Stopped
		case seconds < prevSeconds || prevSeconds == 0:
			m.timerState = Running
		}
		m.lastTimeUpdate = time.Now()

		if seconds > 0 {
			m.alarmTriggered = false
		}
		if seconds == 0 && wasNonZero && !m.alarmTriggered && !m.personalAlarmFired {
			m.alarmTriggered = true
			cmds = append(cmds, playAlarmCmd(), func() tea.Msg {
				return statusMsg("Timer finished!")
			})
		}
		// Personal alarm threshold check
		alarmMins, alarmSecs := 0, 0
		fmt.Sscanf(m.alarmMinsInput, "%d", &alarmMins)
		fmt.Sscanf(m.alarmSecsInput, "%d", &alarmSecs)
		threshold := alarmMins*60 + alarmSecs
		if threshold > 0 {
			if seconds > 0 && seconds <= threshold && !m.personalAlarmFired {
				m.personalAlarmFired = true
				cmds = append(cmds, playAlarmCmd(), func() tea.Msg {
					return statusMsg(fmt.Sprintf("Personal alarm: %d seconds remaining!", seconds))
				})
			} else if seconds > threshold {
				m.personalAlarmFired = false
			}
		}
	default:
	}

	// If the timer was counting down but the update stream has stalled, the
	// shared timer is paused.
	if m.timerState == Running && m.totalSeconds > 0 && time.Since(m.lastTimeUpdate) > 2*time.Second {
		m.timerState = Paused
	}

	select {
	case count := <-m.userCountChan:
		m.SetUserCount(count)
	default:
	}

	select {
	case msg := <-m.messageChan:
		m.SetStatusMessage(msg)
	default:
	}

	if len(cmds) > 0 {
		return tea.Sequence(cmds...)
	}
	return nil
}

func generateRoomID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return "room-" + string(b)
}
