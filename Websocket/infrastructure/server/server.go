package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Primer mensaje debe ser el user_id
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var initData struct {
		UserID int `json:"user_id"`
		DeviceID int `json:"device_id"`
	}
	if err := json.Unmarshal(msg, &initData); err != nil || initData.UserID == 0 {
		conn.Close()
		return
	}

	userID := initData.UserID
	deviceID := initData.DeviceID

	Manager.AddClient(userID,deviceID, conn)
	defer Manager.RemoveClient(userID, deviceID, conn)

	// Mantener conexión viva
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
