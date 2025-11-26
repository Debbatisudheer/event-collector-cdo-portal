package main

import (
	"fmt"
	"math/rand"
	"time"
)

func SimulateFTD(n int) {
	deviceID := fmt.Sprintf("FTD-%d", n)
	RegisterDevice(deviceID, "FTD")
	// Heartbeat sender
go func() {
	for {
		SendHeartbeat(deviceID)
		time.Sleep(5 * time.Second)
	}
}()


	for {
		ev := Event{
			DeviceID: deviceID,
			Device:   "FTD",
			Type:     "telemetry",
			Msg:      "Firewall health status",
			Details:  fmt.Sprintf("cpu=%d mem=%d", rand.Intn(100), rand.Intn(100)),
			Severity: "info",
			Time:     time.Now().Format(time.RFC3339),
			Tenant:   "Tenant-A",
		}

		SendEvent(deviceID, ev)
		time.Sleep(2 * time.Second)
	}
}

func SimulateWSA(n int) {
	deviceID := fmt.Sprintf("WSA-%d", n)
	RegisterDevice(deviceID, "WSA")
	// Heartbeat sender
go func() {
	for {
		SendHeartbeat(deviceID)
		time.Sleep(5 * time.Second)
	}
}()


	for {
		ev := Event{
			DeviceID: deviceID,
			Device:   "WSA",
			Type:     "http_proxy",
			Msg:      "Web request scanned",
			Details:  fmt.Sprintf("url=example.com time=%dms", rand.Intn(300)),
			Severity: "info",
			Time:     time.Now().Format(time.RFC3339),
			Tenant:   "Tenant-A",
		}
		SendEvent(deviceID, ev)
		time.Sleep(3 * time.Second)
	}
}

func SimulateESA(n int) {
	deviceID := fmt.Sprintf("ESA-%d", n)
	RegisterDevice(deviceID, "ESA")
	// Heartbeat sender
go func() {
	for {
		SendHeartbeat(deviceID)
		time.Sleep(5 * time.Second)
	}
}()


	for {
		ev := Event{
			DeviceID: deviceID,
			Device:   "ESA",
			Type:     "email_security",
			Msg:      "Email scanned",
			Details:  fmt.Sprintf("spam_score=%d", rand.Intn(100)),
			Severity: "warning",
			Time:     time.Now().Format(time.RFC3339),
			Tenant:   "Tenant-A",
		}
		SendEvent(deviceID, ev)
		time.Sleep(4 * time.Second)
	}
}

func SimulateISE(n int) {
	deviceID := fmt.Sprintf("ISE-%d", n)
	RegisterDevice(deviceID, "ISE")
	// Heartbeat sender
go func() {
	for {
		SendHeartbeat(deviceID)
		time.Sleep(5 * time.Second)
	}
}()


	for {
		ev := Event{
			DeviceID: deviceID,
			Device:   "ISE",
			Type:     "auth_log",
			Msg:      "User authenticated",
			Details:  fmt.Sprintf("username=user%d", rand.Intn(500)),
			Severity: "info",
			Time:     time.Now().Format(time.RFC3339),
			Tenant:   "Tenant-A",
		}
		SendEvent(deviceID, ev)
		time.Sleep(5 * time.Second)
	}
}
