package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

/*
   CORRELATION ENGINE - SIEM-LIKE RULES

   Rules:
   1) Multiple failed logins (auth_log + "failed") -> brute force
   2) High CPU + error event -> device unhealthy
   3) Event burst (many events in short time) -> possible DoS
   4) Repeated blocked URLs on WSA -> suspicious web activity
*/

type Alert struct {
	DeviceID string `json:"device_id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

var (
	loginFailCount  = make(map[string]int)              // device -> fail count
	eventBurstCount = make(map[string]int)              // device -> events in last 2s
	blockedURLCount = make(map[string]map[string]int)   // device -> url -> count
	alertsByDevice  = make(map[string][]Alert)          // in-memory alert store per device
	correlationMutex sync.Mutex
)

// save alert in memory + print
func saveAlert(alert Alert) {
	correlationMutex.Lock()
	alertsByDevice[alert.DeviceID] = append(alertsByDevice[alert.DeviceID], alert)
	// keep last 50 per device
	if len(alertsByDevice[alert.DeviceID]) > 50 {
		alertsByDevice[alert.DeviceID] = alertsByDevice[alert.DeviceID][len(alertsByDevice[alert.DeviceID])-50:]
	}
	correlationMutex.Unlock()

	fmt.Printf("🚨 ALERT for %s -> %s (%s)\n",
		alert.DeviceID, alert.Message, alert.Severity)
}

// RULE 1: brute force
func detectBruteForce(e Event) {
	if e.Type != "auth_log" {
		return
	}
	if !strings.Contains(strings.ToLower(e.Msg), "failed") {
		return
	}

	correlationMutex.Lock()
	loginFailCount[e.DeviceID]++
	count := loginFailCount[e.DeviceID]
	correlationMutex.Unlock()

	if count >= 5 {
		saveAlert(Alert{
			DeviceID: e.DeviceID,
			Message:  "Multiple failed logins (possible brute force)",
			Severity: "critical",
			Time:     time.Now().Format(time.RFC3339),
		})
		loginFailCount[e.DeviceID] = 0
	}
}

// RULE 2: high CPU + error
func detectHighCPUError(e Event) {
	if e.Type != "error" {
		return
	}
	// try to parse cpu=NN
	cpu := 0
	fmt.Sscanf(e.Details, "cpu=%d", &cpu)
	if cpu >= 90 {
		saveAlert(Alert{
			DeviceID: e.DeviceID,
			Message:  "High CPU with error events (device unhealthy)",
			Severity: "warning",
			Time:     time.Now().Format(time.RFC3339),
		})
	}
}

// RULE 3: DoS (burst of events)
func detectDoS(e Event) {
	correlationMutex.Lock()
	eventBurstCount[e.DeviceID]++
	count := eventBurstCount[e.DeviceID]
	correlationMutex.Unlock()

	if count >= 20 {
		saveAlert(Alert{
			DeviceID: e.DeviceID,
			Message:  "Event burst detected (possible DoS)",
			Severity: "critical",
			Time:     time.Now().Format(time.RFC3339),
		})
		eventBurstCount[e.DeviceID] = 0
	}
}

// reset burst counters every 2 seconds
func startDoSResetter() {
	go func() {
		for {
			time.Sleep(2 * time.Second)
			correlationMutex.Lock()
			eventBurstCount = make(map[string]int)
			correlationMutex.Unlock()
		}
	}()
}

// RULE 4: repeated blocked URLs on WSA
func detectBlockedURL(e Event) {
	if e.Device != "WSA" {
		return
	}
	detailsLower := strings.ToLower(e.Details)
	msgLower := strings.ToLower(e.Msg)

	if !strings.Contains(detailsLower, "blocked") && !strings.Contains(msgLower, "blocked") {
		return
	}

	// simple extraction "url=example.com"
	url := ""
	fmt.Sscanf(e.Details, "url=%s", &url)
	if url == "" {
		return
	}

	correlationMutex.Lock()
	if blockedURLCount[e.DeviceID] == nil {
		blockedURLCount[e.DeviceID] = make(map[string]int)
	}
	blockedURLCount[e.DeviceID][url]++
	count := blockedURLCount[e.DeviceID][url]
	correlationMutex.Unlock()

	if count >= 5 {
		saveAlert(Alert{
			DeviceID: e.DeviceID,
			Message:  "Repeated blocked URL (" + url + ")",
			Severity: "warning",
			Time:     time.Now().Format(time.RFC3339),
		})
		blockedURLCount[e.DeviceID][url] = 0
	}
}

// main entry - called per event
func runCorrelation(e Event) {
	detectBruteForce(e)
	detectHighCPUError(e)
	detectDoS(e)
	detectBlockedURL(e)
}

func init() {
	startDoSResetter()
}
