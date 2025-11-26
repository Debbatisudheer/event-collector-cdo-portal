package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🚀 Event Collector starting on :9000 ...")

	initDB()
	initMetrics()
	startWorkerPool(10)

	// Device ingestion
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/event", eventHandler)
	http.HandleFunc("/ws", wsHandler)

	// NEW — Live Telemetry WebSocket
	http.HandleFunc("/ws/live", liveWSHandler)

	// Dashboard
	http.HandleFunc("/dashboard", dashboardHandler)

	// Search APIs
	http.HandleFunc("/api/events/recent", recentEventsHandler)
	http.HandleFunc("/api/events/by-device-type", eventsByDeviceTypeHandler)
	http.HandleFunc("/api/events/by-device-id", eventsByDeviceIDHandler)
	http.HandleFunc("/api/events/by-severity", eventsBySeverityHandler)
	http.HandleFunc("/api/events/by-tenant", eventsByTenantHandler)
	http.HandleFunc("/api/events/time-range", eventsTimeRangeHandler)

	// Critical alert counter
	http.HandleFunc("/api/alerts/count", alertsCountHandler)
	http.HandleFunc("/api/device/alerts", deviceAlertsHandler)


	http.HandleFunc("/heartbeat", heartbeatHandler)
    http.HandleFunc("/api/health", deviceHealthHandler)
	


	// Prometheus
	http.Handle("/metrics", promHTTPHandler())

	// Start server
	log.Fatal(http.ListenAndServe(":9000", nil))
}
