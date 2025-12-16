package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// SignalMessage represents a WebRTC signaling message exchanged between client and host.
type SignalMessage struct {
	Type      string        `json:"type"` // offer | answer | ice-candidate
	CameraID  string        `json:"cameraId,omitempty"`
	HostID    string        `json:"hostId,omitempty"`
	SDP       *string       `json:"sdp,omitempty"`
	Candidate *ICECandidate `json:"candidate,omitempty"`
}

// ICECandidate mirrors the structure required by browsers for ICE exchange.
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
	SDPMid        string `json:"sdpMid"`
}

// ConnectionManager holds active websocket connections.
type ConnectionManager struct {
	clients map[string]*websocket.Conn // clientId -> conn
	hosts   map[string]*websocket.Conn // hostId -> conn
	mu      sync.RWMutex
}

// Manager is the global singleton instance used by handlers.
var Manager = NewConnectionManager()

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[string]*websocket.Conn),
		hosts:   make(map[string]*websocket.Conn),
	}
}

// Upgrader with permissive CORS for prototype stage.
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (cm *ConnectionManager) RegisterClient(id string, conn *websocket.Conn) {
	cm.mu.Lock()
	cm.clients[id] = conn
	cm.mu.Unlock()
	log.Printf("client connected: %s", id)
}

func (cm *ConnectionManager) RemoveClient(id string) {
	cm.mu.Lock()
	delete(cm.clients, id)
	cm.mu.Unlock()
	log.Printf("client disconnected: %s", id)
}

func (cm *ConnectionManager) RegisterHost(id string, conn *websocket.Conn) {
	cm.mu.Lock()
	cm.hosts[id] = conn
	cm.mu.Unlock()
	log.Printf("host connected: %s", id)
}

func (cm *ConnectionManager) RemoveHost(id string) {
	cm.mu.Lock()
	delete(cm.hosts, id)
	cm.mu.Unlock()
	log.Printf("host disconnected: %s", id)
}

// ForwardToHost forwards a signaling message to a host if connected.
func (cm *ConnectionManager) ForwardToHost(hostID string, msg *SignalMessage) error {
	cm.mu.RLock()
	hConn := cm.hosts[hostID]
	cm.mu.RUnlock()
	if hConn == nil {
		return ErrHostNotFound
	}
	return hConn.WriteJSON(msg)
}

// BroadcastToClients sends a signaling message to all clients (simple prototype logic).
func (cm *ConnectionManager) BroadcastToClients(msg *SignalMessage) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for id, c := range cm.clients {
		if err := c.WriteJSON(msg); err != nil {
			log.Printf("broadcast to client %s failed: %v", id, err)
		}
	}
}

// StartCamera sends a start command to a host.
func (cm *ConnectionManager) StartCamera(hostID, cameraID string) error {
	cm.mu.RLock()
	hConn := cm.hosts[hostID]
	cm.mu.RUnlock()
	if hConn == nil {
		return ErrHostNotFound
	}
	cmd := map[string]any{"command": "start_stream", "cameraId": cameraID}
	return hConn.WriteJSON(cmd)
}

// StopCamera sends a stop command to a host.
func (cm *ConnectionManager) StopCamera(hostID, cameraID string) error {
	cm.mu.RLock()
	hConn := cm.hosts[hostID]
	cm.mu.RUnlock()
	if hConn == nil {
		return ErrHostNotFound
	}
	cmd := map[string]any{"command": "stop_stream", "cameraId": cameraID}
	return hConn.WriteJSON(cmd)
}

// Minimal errors for prototype.
var (
	ErrHostNotFound   = websocket.ErrBadHandshake // reuse existing error type for brevity
	ErrClientNotFound = websocket.ErrBadHandshake
)

// HandleClientLoop reads messages from a client and forwards to host when HostID present.
func (cm *ConnectionManager) HandleClientLoop(clientID string, conn *websocket.Conn) {
	defer conn.Close()
	for {
		var msg SignalMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("client read err: %v", err)
			break
		}
		if msg.HostID != "" {
			if err := cm.ForwardToHost(msg.HostID, &msg); err != nil {
				log.Printf("forward to host failed: %v", err)
			}
		}
	}
	cm.RemoveClient(clientID)
}

// HandleHostLoop reads messages from a host and broadcasts to clients.
func (cm *ConnectionManager) HandleHostLoop(hostID string, conn *websocket.Conn) {
	defer conn.Close()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("host read err: %v", err)
			break
		}
		var msg SignalMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("unmarshal host message err: %v", err)
			continue
		}
		cm.BroadcastToClients(&msg)
	}
	cm.RemoveHost(hostID)
}
