package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/bchanona/websocket_backend/Websocket/domain"
	"github.com/gorilla/websocket"
)

// ClientManager gestiona las conexiones WebSocket activas
type ClientManager struct {
	mu       sync.Mutex
    clients  map[int]map[*websocket.Conn]bool // userID → conexiones
}

var Manager = ClientManager{
	 clients: make(map[int]map[*websocket.Conn]bool),
}

func (cm *ClientManager) AddClient(userID int, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if cm.clients[userID] == nil {
        cm.clients[userID] = make(map[*websocket.Conn]bool)
    }
    cm.clients[userID][conn] = true
}


// Remover un cliente WebSocket
func (cm *ClientManager) RemoveClient(userID int, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if cm.clients[userID] != nil {
        delete(cm.clients[userID], conn)
        if len(cm.clients[userID]) == 0 {
            delete(cm.clients, userID)
        }
    }
}


func (cm *ClientManager) SendToUser(userID int, message domain.Message) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    conns, ok := cm.clients[userID]
    if !ok {
        return // usuario no conectado
    }

    jsonMsg, err := json.Marshal(message)
    if err != nil {
        log.Println("Error serializing message:", err)
        return
    }

    for conn := range conns {
        if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
            log.Println("Error sending message:", err)
            conn.Close()
            delete(conns, conn)
        }
    }
}
