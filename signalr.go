package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SignalRClient struct {
	conn         *websocket.Conn
	roomID       string
	baseURL      string
	messageID    int
	messageMutex sync.Mutex
	handlers     map[string]func([]interface{})
	closeChan    chan struct{}
	connected    bool
	connMutex    sync.RWMutex
}

type HandshakeRequest struct {
	Protocol string `json:"protocol"`
	Version  int    `json:"version"`
}

type HandshakeResponse struct {
	Error string `json:"error"`
}

func NewSignalRClient(baseURL, roomID string) *SignalRClient {
	return &SignalRClient{
		baseURL:   baseURL,
		roomID:    roomID,
		messageID: 0,
		handlers:  make(map[string]func([]interface{})),
		closeChan: make(chan struct{}),
		connected: false,
	}
}

type negotiateResponse struct {
	ConnectionToken string `json:"connectionToken"`
}

func (c *SignalRClient) negotiate() (string, error) {
	negotiateURL := c.baseURL + "/timerHub/negotiate?negotiateVersion=1"
	resp, err := http.Post(negotiateURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("negotiate request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read negotiate response: %w", err)
	}

	var nr negotiateResponse
	if err := json.Unmarshal(body, &nr); err != nil {
		return "", fmt.Errorf("parse negotiate response: %w", err)
	}
	if nr.ConnectionToken == "" {
		return "", fmt.Errorf("empty connection token from negotiate")
	}
	return nr.ConnectionToken, nil
}

func (c *SignalRClient) Connect() error {
	conn, err := c.dial()
	if err != nil {
		return err
	}

	c.connMutex.Lock()
	c.conn = conn
	c.connected = true
	c.connMutex.Unlock()

	go c.readMessages()

	return c.joinRoom()
}

func (c *SignalRClient) dial() (*websocket.Conn, error) {
	token, err := c.negotiate()
	if err != nil {
		return nil, fmt.Errorf("negotiate: %w", err)
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	wsURL := url.URL{
		Scheme: func() string {
			if u.Scheme == "https" {
				return "wss"
			}
			return "ws"
		}(),
		Host:     u.Host,
		Path:     "/timerHub",
		RawQuery: "id=" + url.QueryEscape(token),
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial websocket: %w", err)
	}

	handshakeData, _ := json.Marshal(HandshakeRequest{Protocol: "json", Version: 1})
	if err := conn.WriteMessage(websocket.TextMessage, append(handshakeData, 0x1E)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send handshake: %w", err)
	}

	_, response, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read handshake response: %w", err)
	}

	respStr := strings.TrimRight(string(response), "\x1e")
	var handshakeResp HandshakeResponse
	if err := json.Unmarshal([]byte(respStr), &handshakeResp); err == nil && handshakeResp.Error != "" {
		conn.Close()
		return nil, fmt.Errorf("handshake error: %s", handshakeResp.Error)
	}

	return conn, nil
}

func (c *SignalRClient) joinRoom() error {
	args := []interface{}{c.roomID}
	return c.invoke("JoinSession", args)
}

func (c *SignalRClient) invoke(method string, args []interface{}) error {
	c.connMutex.RLock()
	if !c.connected || c.conn == nil {
		c.connMutex.RUnlock()
		return fmt.Errorf("not connected")
	}
	conn := c.conn
	c.connMutex.RUnlock()

	c.messageMutex.Lock()
	c.messageID++
	id := c.messageID
	c.messageMutex.Unlock()

	msg := map[string]interface{}{
		"type":      1,
		"target":    method,
		"arguments": args,
		"invocationId": fmt.Sprintf("%d", id),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal invoke: %w", err)
	}

	return conn.WriteMessage(websocket.TextMessage, append(data, 0x1E))
}

func (c *SignalRClient) RegisterHandler(target string, handler func([]interface{})) {
	c.handlers[target] = handler
}

func (c *SignalRClient) readMessages() {
	defer func() {
		c.connMutex.Lock()
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		c.connMutex.Unlock()
		close(c.closeChan)
	}()

	for {
		c.connMutex.RLock()
		if !c.connected || c.conn == nil {
			c.connMutex.RUnlock()
			return
		}
		conn := c.conn
		c.connMutex.RUnlock()

		_, message, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

		c.processMessage(message)
	}
}

func (c *SignalRClient) processMessage(data []byte) {
	// A single WebSocket frame may contain multiple 0x1E-delimited records.
	records := strings.Split(string(data), "\x1e")
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		c.processRecord([]byte(record))
	}
}

func (c *SignalRClient) processRecord(data []byte) {
	if len(data) == 0 {
		return
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to parse record: %v", err)
		return
	}

	msgType, ok := msg["type"].(float64)
	if !ok {
		return
	}

	switch int(msgType) {
	case 1:
		c.handleInvocation(msg)
	case 2:
		c.handleStreamItem(msg)
	case 3:
		c.handleCompletion(msg)
	case 4:
		c.handleStreamInvocation(msg)
	case 5:
		c.handleCancelInvocation(msg)
	case 6:
		c.handlePing()
	case 7:
		c.handleClose(msg)
	}
}

func (c *SignalRClient) handleInvocation(msg map[string]interface{}) {
	target, _ := msg["target"].(string)
	arguments, _ := msg["arguments"].([]interface{})

	if handler, exists := c.handlers[target]; exists {
		handler(arguments)
	}
}

func (c *SignalRClient) handleStreamItem(msg map[string]interface{}) {
}

func (c *SignalRClient) handleCompletion(msg map[string]interface{}) {
}

func (c *SignalRClient) handleStreamInvocation(msg map[string]interface{}) {
}

func (c *SignalRClient) handleCancelInvocation(msg map[string]interface{}) {
}

func (c *SignalRClient) handlePing() {
}

func (c *SignalRClient) handleClose(msg map[string]interface{}) {
	c.connMutex.Lock()
	c.connected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMutex.Unlock()
}

func (c *SignalRClient) StartTimer(minutes, seconds int) error {
	args := []interface{}{c.roomID, minutes, seconds}
	return c.invoke("StartTimer", args)
}

func (c *SignalRClient) TogglePause() error {
	args := []interface{}{c.roomID}
	return c.invoke("TogglePause", args)
}

func (c *SignalRClient) ResetTimer() error {
	args := []interface{}{c.roomID}
	return c.invoke("ResetTimer", args)
}

func (c *SignalRClient) AdjustTime(deltaSeconds int) error {
	args := []interface{}{c.roomID, deltaSeconds}
	return c.invoke("AdjustTime", args)
}

func (c *SignalRClient) Close() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if !c.connected || c.conn == nil {
		return nil
	}

	c.connected = false
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *SignalRClient) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.connected
}

func decodeHandshake(data string) (map[string]interface{}, error) {
	parts := strings.Split(data, ";")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid handshake format")
	}

	result := make(map[string]interface{})
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result, nil
}

func parseSignalRMessage(data string) (map[string]interface{}, error) {
	var msg map[string]interface{}
	err := json.Unmarshal([]byte(data), &msg)
	return msg, err
}

func decodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func parseInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert to int")
	}
}
