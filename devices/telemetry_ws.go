package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func StartTelemetryWebSocket(deviceID string) {
	token := Tokens[deviceID]
	if token == "" {
		fmt.Println("No token for", deviceID)
		return
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:9000/ws", header)
	if err != nil {
		fmt.Println("WS connect failed for", deviceID, ":", err)
		return
	}

	go func() {
		defer conn.Close()
		for {
			e := Event{
				DeviceID: deviceID,
				Type:     "telemetry",
				Msg:      "health",
				Details:  fmt.Sprintf("cpu=%d mem=%d disk=%d", rand.Intn(100), rand.Intn(100), rand.Intn(100)),
				Severity: "info",
				Time:     time.Now().Format(time.RFC3339),
			}

			data, _ := json.Marshal(e)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("WS send failed for", deviceID, ":", err)
				return
			}

			time.Sleep(2 * time.Second)
		}
	}()
}
