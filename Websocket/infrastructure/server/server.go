package server

import (
	"log"
	"net/http"

	"github.com/bchanona/websocket_backend/Websocket/application"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Manejador de conexiones WebSocket
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	application.Manager.AddClient(conn)
	defer application.Manager.RemoveClient(conn)

	// Mantener la conexión abierta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}