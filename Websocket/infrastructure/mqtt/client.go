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

	// Enviar por WebSocket
	application.Manager.Broadcast(payload)

	var apiURL string
	var data map[string]interface{}

	
	if payload.Spo2 != 0 {
		apiURL = apiVitals + "/oxygen/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Spo2,
			"device_id":   payload.DeviceId,
		}
	} else if payload.Bpm != 0 {
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
			"measurement": payload.Bpm2,
			"device_id":   payload.DeviceId,
		}
	} else if payload.Temperature != 0 {
		apiURL = apiVitals + "/temperature/"
		data = map[string]interface{}{
			"user_id":     payload.UserID,
			"measurement": payload.Temperature,
			"device_id":   payload.DeviceId,
		}
	}

	// Validación de condiciones anormales
	isAnormal := false
	var mensaje string

	if payload.Temperature > 37.5 {
		isAnormal = true
		mensaje += fmt.Sprintf("Temperatura alta: %.1f°C. ", payload.Temperature)
	}
	if payload.Bpm < 60 || payload.Bpm > 100 {
		isAnormal = true
		mensaje += fmt.Sprintf("Ritmo cardíaco anormal: %d bpm. ", payload.Bpm)
	}
	if payload.Spo2 < 95 {
		isAnormal = true
		mensaje += fmt.Sprintf("Oxigenación baja: %d%%. ", payload.Spo2)
	}

	if isAnormal {
		notifData := map[string]interface{}{
			"user_id": payload.UserID,
			"body":    mensaje,
			"reading": false, 
		}

		notifJSON, err := json.Marshal(notifData)
		if err != nil {
			fmt.Println("Error al convertir notificación a JSON:", err)
		} else {
			resp, err := http.Post("http://localhost:8081/user/saveNotification", "application/json", bytes.NewBuffer(notifJSON))
			if err != nil {
				fmt.Println("Error al enviar notificación:", err)
			} else {
				defer resp.Body.Close()
				fmt.Println("Notificación enviada con status:", resp.Status)
			}
		}
	}

	// Enviar datos a la API de vitales si hay URL definida
	if apiURL != "" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error al convertir data a JSON:", err)
			return
		}

		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error al enviar datos a la API:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("API response: %s\n", resp.Status)
	} else {
		fmt.Println("No valid data was found to send to the API")
	}

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
