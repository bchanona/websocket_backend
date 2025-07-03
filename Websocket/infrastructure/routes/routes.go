package routes

import (
	"net/http"

	"github.com/bchanona/websocket_backend/Websocket/infrastructure/server"
)

func SetupRoutes() {
	http.HandleFunc("/ws", server.WSHandler)
}