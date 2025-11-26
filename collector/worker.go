package main

import (
	"fmt"
	"time"
)

var eventChan = make(chan Event, 2000)

func worker(id int) {
	for e := range eventChan {

		// 1 — Prometheus metrics
		recordMetrics(e)

		// 2 — Critical alert detection
		checkCritical(e)

		// 3 — SIEM correlation engine (bruteforce, DoS, WSA blocked URLs, CPU errors)
		runCorrelation(e)

		// 4 — Logs for debugging
		switch e.Type {

		case "error":
			fmt.Printf("❌ [W%d] ERROR from %s -> %s (%s)\n",
				id, e.DeviceID, e.Msg, e.Severity)

		case "telemetry":
			fmt.Printf("📊 [W%d] TELEMETRY from %s -> %s\n",
				id, e.DeviceID, e.Details)

			// (Optional later) Live telemetry broadcast
			// broadcastTelemetry(e)

		default:
			fmt.Printf("[W%d] [%s] %s -> %s (%s)\n",
				id, e.Device, e.Type, e.Msg, e.Severity)
			time.Sleep(20 * time.Millisecond)
		}

		// 5 — Store event into BoltDB with indexes
		saveEventToDB(e)
	}
}

func startWorkerPool(n int) {
	for i := 1; i <= n; i++ {
		go worker(i)
	}
}
