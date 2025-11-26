package main

import (
	"encoding/json"
	"net/http"
)

// returns all alerts for a given device from in-memory map
func deviceAlertsHandler(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	correlationMutex.Lock()
	alerts := alertsByDevice[deviceID]
	correlationMutex.Unlock()

	if alerts == nil {
		alerts = []Alert{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(alerts)
}
