package mqtt

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/bchanona/websocket_backend/Websocket/domain"
    "github.com/bchanona/websocket_backend/Websocket/infrastructure/server"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Job struct {
    URL  string
    Body []byte
}


var (
    jobQueue   chan Job
    httpClient *http.Client

    minWorkers = 20
    maxWorkers = 200

    scaleUpThreshold = 60 

    workers int
    mu      sync.Mutex

    // Circuit Breaker
    failures        = 0
    openCircuit     = false
    circuitOpenedAt time.Time

    // Rate limiting
    limiter = time.Tick(15 * time.Millisecond) 
)

// --------------------- INIT ---------------------

func InitDynamicPool(queueSize int) {
    jobQueue = make(chan Job, queueSize)

    httpClient = &http.Client{
        Timeout: 10 * time.Second,
    }

    workers = minWorkers
    for i := 0; i < minWorkers; i++ {
        go worker()
    }

    go autoscaler()
}


func worker() {
    idle := time.NewTimer(10 * time.Second)

    for {
        select {

        case job := <-jobQueue:
            idle.Reset(10 * time.Second)

            <-limiter

       
            if openCircuit && time.Since(circuitOpenedAt) < 5*time.Second {

                jobQueue <- job
                continue
            } else if openCircuit {

                openCircuit = false
                failures = 0
            }

      
            req, err := http.NewRequest("POST", job.URL, bytes.NewBuffer(job.Body))
            if err != nil {
                fmt.Println("error request:", err)
                continue
            }

            req.Header.Set("Content-Type", "application/json")

            resp, err := httpClient.Do(req)
            if err != nil {
                fmt.Println("error POST:", err)

                failures++
                if failures >= 5 {
                    openCircuit = true
                    circuitOpenedAt = time.Now()
                }
                continue
            }

            resp.Body.Close()
            failures = 0 

        case <-idle.C:
            mu.Lock()
            if workers > minWorkers {
                workers--
                mu.Unlock()
                return
            }
            mu.Unlock()
            idle.Reset(10 * time.Second)
        }
    }
}



func autoscaler() {
    for {
        qlen := len(jobQueue)

        mu.Lock()

        // escala solo si cola grande + API sana
        if qlen > scaleUpThreshold && workers < maxWorkers && !openCircuit {
            extra := (qlen / scaleUpThreshold)
            if extra < 1 {
                extra = 1
            }
            if workers+extra > maxWorkers {
                extra = maxWorkers - workers
            }

            for i := 0; i < extra; i++ {
                workers++
                go worker()
            }
        }

        mu.Unlock()
        time.Sleep(2 * time.Second)
    }
}

// --------------------- MQTT HANDLER ---------------------

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

    apiVitals := os.Getenv("Api_Vitals")
    apiUser := os.Getenv("Api_User")

    var payload domain.Message
    if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
        fmt.Println("Error al decodificar mensaje:", err)
        return
    }

    go server.Manager.SendToUserDevice(payload.UserID, payload.DeviceId, payload)

    var apiURL string
    var data map[string]interface{}

    switch {
    case payload.Spo2 != 0:
        apiURL = apiVitals + "/oxygen/"
        data = map[string]interface{}{
            "user_id":    payload.UserID,
            "measurement": payload.Spo2,
            "device_id":  payload.DeviceId,
        }

    case payload.Bpm != 0:
        apiURL = apiVitals + "/heartRate/"
        data = map[string]interface{}{
            "user_id":    payload.UserID,
            "measurement": payload.Bpm,
            "device_id":  payload.DeviceId,
        }

    case payload.Bpm2 != 0:
        apiURL = apiVitals + "/heartRate/"
        data = map[string]interface{}{
            "user_id":    payload.UserID,
            "measurement": payload.Bpm2,
            "device_id":  payload.DeviceId,
        }

    case payload.Temperature != 0:
        apiURL = apiVitals + "/temperature/"
        data = map[string]interface{}{
            "user_id":    payload.UserID,
            "measurement": payload.Temperature,
            "device_id":  payload.DeviceId,
        }
    }

    // VALIDACIONES
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
        notif := map[string]interface{}{
            "user_id": payload.UserID,
            "body":    mensaje,
            "reading": false,
        }
        notifJSON, _ := json.Marshal(notif)

        jobQueue <- Job{
            URL:  apiUser + "/user/saveNotification",
            Body: notifJSON,
        }
    }

    if apiURL != "" {
        jsonData, _ := json.Marshal(data)
        jobQueue <- Job{URL: apiURL, Body: jsonData}
    }
}
