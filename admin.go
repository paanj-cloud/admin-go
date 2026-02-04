package admin

// Test sync: 2026-02-04

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// AdminHttpClient
type AdminHttpClient struct {
	secretKey string
	apiUrl    string
	client    *http.Client
}

func NewAdminHttpClient(secretKey, apiUrl string) *AdminHttpClient {
	return &AdminHttpClient{
		secretKey: secretKey,
		apiUrl:    apiUrl,
		client:    &http.Client{},
	}
}

func (c *AdminHttpClient) Request(method, path string, body interface{}) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.apiUrl+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.secretKey) // JS SDK uses X-API-Key for admin secret

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error: status=%d body=%s", resp.StatusCode, string(responseBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil
	}

	return result, nil
}

// AdminWebSocketClient
type AdminWebSocketClient struct {
	secretKey            string
	wsUrl                string
	conn                 *websocket.Conn
	mu                   sync.Mutex
	isConnected          bool
	eventHandlers        map[string][]func(interface{})
	autoReconnect        bool
	reconnectInterval    time.Duration
	maxReconnectAttempts int
}

func NewAdminWebSocketClient(secretKey, wsUrl string, autoReconnect bool, interval time.Duration, maxAttempts int) *AdminWebSocketClient {
	return &AdminWebSocketClient{
		secretKey:            secretKey,
		wsUrl:                wsUrl,
		autoReconnect:        autoReconnect,
		reconnectInterval:    interval,
		maxReconnectAttempts: maxAttempts,
		eventHandlers:        make(map[string][]func(interface{})),
	}
}

func (c *AdminWebSocketClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	// Match JS SDK format: /ws/admin?secretKey=...
	url := fmt.Sprintf("%s/ws/admin?secretKey=%s", c.wsUrl, c.secretKey)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.isConnected = true

	go c.listen()

	return nil
}

func (c *AdminWebSocketClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.isConnected = false
	}
}

func (c *AdminWebSocketClient) listen() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
		if c.autoReconnect {
			time.Sleep(c.reconnectInterval)
			c.Connect()
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("ws read error:", err)
			return
		}

		var eventData struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}

		if err := json.Unmarshal(message, &eventData); err != nil {
			continue
		}

		c.mu.Lock()
		handlers := c.eventHandlers[eventData.Type]
		c.mu.Unlock()

		for _, handler := range handlers {
			go handler(eventData.Data)
		}
	}
}

func (c *AdminWebSocketClient) Send(data interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return fmt.Errorf("websocket not connected")
	}

	return c.conn.WriteJSON(data)
}

func (c *AdminWebSocketClient) On(event string, callback func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers[event] = append(c.eventHandlers[event], callback)
}

func (c *AdminWebSocketClient) IsConnectedStatus() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

func (c *AdminWebSocketClient) Subscribe(subscription interface{}) error {
	return c.Send(map[string]interface{}{
		"type": "subscribe",
		"data": subscription,
	})
}

// PaanjAdmin
type AdminOptions struct {
	ApiUrl               string
	WsUrl                string
	AutoReconnect        bool
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
}

type PaanjAdmin struct {
	secretKey  string
	wsClient   *AdminWebSocketClient
	httpClient *AdminHttpClient
	options    AdminOptions
}

func NewAdmin(secretKey string, options AdminOptions) *PaanjAdmin {
	params := options
	if params.ApiUrl == "" {
		params.ApiUrl = "http://localhost:3000"
	}
	if params.WsUrl == "" {
		params.WsUrl = "ws://localhost:8090"
	}
	if params.ReconnectInterval == 0 {
		params.ReconnectInterval = 5 * time.Second
	}
	if params.MaxReconnectAttempts == 0 {
		params.MaxReconnectAttempts = 10
	}

	admin := &PaanjAdmin{
		secretKey: secretKey,
		options:   params,
	}

	admin.wsClient = NewAdminWebSocketClient(
		secretKey,
		params.WsUrl,
		params.AutoReconnect,
		params.ReconnectInterval,
		params.MaxReconnectAttempts,
	)

	admin.httpClient = NewAdminHttpClient(secretKey, params.ApiUrl)

	return admin
}

func (c *PaanjAdmin) Connect() error {
	return c.wsClient.Connect()
}

func (c *PaanjAdmin) Disconnect() {
	c.wsClient.Disconnect()
}

func (c *PaanjAdmin) IsConnected() bool {
	return c.wsClient.IsConnectedStatus()
}

func (c *PaanjAdmin) GetWebSocket() *AdminWebSocketClient {
	return c.wsClient
}

func (c *PaanjAdmin) GetHttpClient() *AdminHttpClient {
	return c.httpClient
}

func (c *PaanjAdmin) Subscribe(subscription interface{}) error {
	return c.wsClient.Subscribe(subscription)
}

func (c *PaanjAdmin) On(event string, callback func(interface{})) {
	c.wsClient.On(event, callback)
}
