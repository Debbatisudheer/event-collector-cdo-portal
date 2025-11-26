package main

import (
	"encoding/json"
	"net/http"
)

func alertsCountHandler(w http.ResponseWriter, r *http.Request) {
	alertMu.Lock()
	count := criticalAlerts
	alertMu.Unlock()

	json.NewEncoder(w).Encode(map[string]int{
		"critical_alerts": count,
	})
}
