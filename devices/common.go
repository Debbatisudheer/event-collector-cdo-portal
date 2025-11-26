package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Shared Event struct across all device types
type Event struct {
	DeviceID string `json:"device_id"`
	Device   string `json:"device"`
	Type     string `json:"type"`
	Msg      string `json:"msg"`
	Details  string `json:"details"`
	Severity string `json:"severity"`
	Time     string `json:"time"`
	Tenant   string `json:"tenant,omitempty"`
}

type RegisterRequest struct {
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`
	Tenant   string `json:"tenant"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

// store tokens per-device
var Tokens = make(map[string]string)

func RegisterDevice(deviceID, devType string) string {
	req := RegisterRequest{
		DeviceID: deviceID,
		Type:     devType,
		Tenant:   "Tenant-A",
	}

	data, _ := json.Marshal(req)

	resp, err := http.Post("http://localhost:9000/register",
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("❌ Registration failed:", err)
		return ""
	}
	defer resp.Body.Close()

	var regResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		fmt.Println("❌ Failed to decode registration response:", err)
		return ""
	}

	fmt.Printf("🔑 Device %s registered, token=%s\n", deviceID, regResp.Token)
	Tokens[deviceID] = regResp.Token
	return regResp.Token
}

// SEND EVENT via HTTP
func SendEvent(deviceID string, event Event) {
	token, ok := Tokens[deviceID]
	if !ok {
		fmt.Println("❗ No token for device:", deviceID)
		return
	}

	jsonBytes, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", "http://localhost:9000/event",
		bytes.NewBuffer(jsonBytes))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Failed to send event:", err)
	}
}

// SEND HEARTBEAT every 5 sec
func SendHeartbeat(deviceID string) {
	token := Tokens[deviceID]

	req, _ := http.NewRequest("POST", "http://localhost:9000/heartbeat", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 2 * time.Second}
	client.Do(req)
}
