package server

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bchanona/websocket_backend/Websocket/domain"
	"github.com/gorilla/websocket"
)

// ClientManager gestiona las conexiones WebSocket activas
type ClientManager struct {
    mu      sync.Mutex
    clients map[int]map[int]map[*websocket.Conn]bool
}

var Manager = ClientManager{
    clients: make(map[int]map[int]map[*websocket.Conn]bool),
}

func (cm *ClientManager) AddClient(userID, deviceID int, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if cm.clients[userID] == nil {
        cm.clients[userID] = make(map[int]map[*websocket.Conn]bool)
    }
    if cm.clients[userID][deviceID] == nil {
        cm.clients[userID][deviceID] = make(map[*websocket.Conn]bool)
    }

    cm.clients[userID][deviceID][conn] = true
}



// Remover un cliente WebSocket
func (cm *ClientManager) RemoveClient(userID, deviceID int, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if cm.clients[userID] != nil && cm.clients[userID][deviceID] != nil {
        delete(cm.clients[userID][deviceID], conn)

        if len(cm.clients[userID][deviceID]) == 0 {
            delete(cm.clients[userID], deviceID)
        }
        if len(cm.clients[userID]) == 0 {
            delete(cm.clients, userID)
        }
    }
}



func (cm *ClientManager) SendToUserDevice(userID, deviceID int, msg domain.Message) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    devices, ok := cm.clients[userID]
    if !ok {
        // Usuario no conectado
        fmt.Printf("[WS] No hay conexiones para user_id=%d\n", userID)
        return
    }

    conns, ok := devices[deviceID]
    if !ok {
        // Usuario conectado, pero no ese dispositivo
        fmt.Printf("[WS] No hay conexiones para user_id=%d device_id=%d\n", userID, deviceID)
        return
    }

    jsonMsg, err := json.Marshal(msg)
    if err != nil {
        fmt.Println("[WS] Error al serializar mensaje:", err)
        return
    }

    for conn := range conns {
        if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
            fmt.Printf("[WS] Error enviando a user_id=%d device_id=%d: %v\n", userID, deviceID, err)
            conn.Close()
            delete(conns, conn)
        } else {
            fmt.Printf("[WS] Mensaje enviado a user_id=%d device_id=%d\n", userID, deviceID)
        }
    }
}



