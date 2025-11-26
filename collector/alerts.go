package main

import "fmt"

// Simple placeholder email alert system.
// Not using real SMTP to avoid configuration issues.

func sendCriticalEmail(e Event) {
    fmt.Printf("📧 [ALERT] CRITICAL ERROR from %s: %s (%s)\n",
        e.DeviceID, e.Msg, e.Details)
}
