package main

import (
	"log"
	"net/http"

	"github.com/bchanona/websocket_backend/Websocket/infrastructure/mqtt"
	"github.com/bchanona/websocket_backend/Websocket/infrastructure/routes"
)

func main() {
	routes.SetupRoutes()

	go mqtt.StartMQTTClient()

	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}