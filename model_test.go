package main

import (
	"testing"
)

func TestTimerModelFormatTimer(t *testing.T) {
	model := NewTimerModel("https://knix.ovh", "test-room")

	model.SetTime(0)
	if model.FormatTimer() != "00:00" {
		t.Errorf("Expected 00:00, got %s", model.FormatTimer())
	}

	model.SetTime(5)
	if model.FormatTimer() != "00:05" {
		t.Errorf("Expected 00:05, got %s", model.FormatTimer())
	}

	model.SetTime(65)
	if model.FormatTimer() != "01:05" {
		t.Errorf("Expected 01:05, got %s", model.FormatTimer())
	}

	model.SetTime(3600)
	if model.FormatTimer() != "60:00" {
		t.Errorf("Expected 60:00, got %s", model.FormatTimer())
	}
}

func TestTimerModelState(t *testing.T) {
	model := NewTimerModel("https://knix.ovh", "test-room")

	if model.connectionState != Disconnected {
		t.Errorf("Expected Disconnected, got %v", model.connectionState)
	}

	model.SetConnectionState(Connected)
	if model.connectionState != Connected {
		t.Errorf("Expected Connected, got %v", model.connectionState)
	}

	if model.ConnectionStateText() != "Connected" {
		t.Errorf("Expected 'Connected', got %s", model.ConnectionStateText())
	}
}

func TestTimerModelTimerState(t *testing.T) {
	model := NewTimerModel("https://knix.ovh", "test-room")

	if model.timerState != Stopped {
		t.Errorf("Expected Stopped, got %v", model.timerState)
	}

	model.SetTimerState(Running)
	if model.timerState != Running {
		t.Errorf("Expected Running, got %v", model.timerState)
	}

	if model.TimerStateText() != "Running" {
		t.Errorf("Expected 'Running', got %s", model.TimerStateText())
	}

	model.SetTimerState(Paused)
	if model.TimerStateText() != "Paused" {
		t.Errorf("Expected 'Paused', got %s", model.TimerStateText())
	}
}

func TestGenerateRoomID(t *testing.T) {
	roomID := generateRoomID()
	if len(roomID) != 13 {
		t.Errorf("Expected room ID length 13, got %d", len(roomID))
	}

	if roomID[:5] != "room-" {
		t.Errorf("Expected room ID to start with 'room-', got %s", roomID[:5])
	}
}

func TestSignalRClientCreation(t *testing.T) {
	client := NewSignalRClient("https://example.com", "test-room")
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.baseURL != "https://example.com" {
		t.Errorf("Expected base URL 'https://example.com', got %s", client.baseURL)
	}

	if client.roomID != "test-room" {
		t.Errorf("Expected room ID 'test-room', got %s", client.roomID)
	}

	if client.IsConnected() {
		t.Error("Expected client to be disconnected initially")
	}
}
