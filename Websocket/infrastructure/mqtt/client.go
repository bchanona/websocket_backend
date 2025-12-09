package mqtt

import (
	"fmt"
	"log"
	"os"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

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
