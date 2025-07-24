package domain

type Message struct {
	DeviceId    int     `json:"device_id"`
	UserID      int     `json:"user_id"`
	Bpm      int `json:"bpm"`
	Spo2     int `json:"spo2"`
	Bpm2     int `json:"bpm2"`
	Moving   bool  `json:"moving"`
	Temperature float64 `json:"temperature"`
}