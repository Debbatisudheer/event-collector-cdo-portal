package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	deviceLastSeen = make(map[string]time.Time)
	dsMu           sync.Mutex
)

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	dev, ok := validateToken(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dsMu.Lock()
	deviceLastSeen[dev.ID] = time.Now()
	dsMu.Unlock()

	w.Write([]byte("ok"))
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	status := make(map[string]string)

	dsMu.Lock()
	for devID, last := range deviceLastSeen {
		if now.Sub(last) > 15*time.Second {
			status[devID] = "offline"
		} else {
			status[devID] = "online"
		}
	}
	dsMu.Unlock()

	json.NewEncoder(w).Encode(status)
}
