package mqtt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bchanona/websocket_backend/Websocket/application"
	"github.com/bchanona/websocket_backend/Websocket/domain"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// Procesador de mensajes MQTT
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	apiVitals := os.Getenv("Api_Vitals")
	payloadStr := string(msg.Payload())

	var payload domain.Message

	// Decodificar el JSON del mensaje recibido
	err := json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		fmt.Println("Error al decodificar el mensaje JSON:", err)
		return
	}

	application.Manager.Broadcast(payload)

	var apiURL string
	var data map[string]interface{}

	if payload.Spo2 != 0 { // Si hay un valor de SPO2, se enviará a la ruta de oxígeno
		apiURL = apiVitals + "/oxygen/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Spo2,
			"device_id":   payload.DeviceId,
		}
	} else if payload.Bpm != 0 { // Si hay un valor de BPM, se enviará a la ruta de frecuencia cardíaca
		apiURL = apiVitals + "/heartRate/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Bpm,
			"device_id":   payload.DeviceId,
		}
	} else if payload.Bpm2 != 0 {

		apiURL = apiVitals + "/heartRate/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Bpm,
			"device_id":   payload.DeviceId,
		}
	} else if payload.Temperature != 0 { // Si hay un valor de temperatura, se enviará a la ruta de temperatura
		apiURL = apiVitals + "/temperature/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Temperature,
			"device_id":   payload.DeviceId,
		}
	}
	// Si se asignó una URL, hacer la solicitud HTTP POST
	if apiURL != "" {
		// Convertir los datos a JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error converting data to JSON:", err)
			return
		}

		// Hacer la solicitud HTTP POST a la API
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			return
		}
		defer resp.Body.Close()

		// Imprimir la respuesta de la API (opcional)
		fmt.Printf("API response: %s\n", resp.Status)
	} else {
		fmt.Println("No valid data was found to send to the API")
	}

	// Enviar el mensaje a todos los clientes WebSocket

	fmt.Printf("Message received in [%s]: %s\n", msg.Topic(), payloadStr)
}

// Iniciar la conexión MQTT y suscribirse a tópicos
func StartMQTTClient() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	userName := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	broker := os.Getenv("RABBITMQ_IP")
	clientID := os.Getenv("RABBITMQ_CLIENT_ID")
	topic := os.Getenv("TOPIC")

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://" + broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetUsername(userName)
	opts.SetPassword(password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	if token := client.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	} else {
		fmt.Println("Suscrito al tópico:", topic)
	}

}
