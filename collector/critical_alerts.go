package main

import (
	"fmt"
	"strings"
	"sync"
)

var (
	criticalAlerts = 0
	alertMu        sync.Mutex
)

func checkCritical(e Event) {
	// Severity check
	if strings.ToLower(e.Severity) == "critical" {
		incrementAlert()
		return
	}

	// Keyword check
	s := strings.ToLower(e.Msg + " " + e.Details)
	if strings.Contains(s, "error") ||
		strings.Contains(s, "fail") ||
		strings.Contains(s, "panic") ||
		strings.Contains(s, "drop") {
		incrementAlert()
		return
	}

	// CPU/MEM threshold check
	if parseValue(e.Details, "cpu") > 90 {
		incrementAlert()
		return
	}

	if parseValue(e.Details, "mem") > 90 {
		incrementAlert()
		return
	}
}

func incrementAlert() {
	alertMu.Lock()
	criticalAlerts++
	alertMu.Unlock()
}

func parseValue(details, key string) int {
	var val int
	_, err := fmt.Sscanf(details, key+"=%d", &val)
	if err != nil {
		return 0
	}
	return val
}
