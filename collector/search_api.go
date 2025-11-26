package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func parseLimit(r *http.Request) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return 50
	}
	n, err := strconv.Atoi(q)
	if err != nil || n <= 0 {
		return 50
	}
	if n > 500 {
		return 500
	}
	return n
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// GET /api/events/recent?limit=100
func recentEventsHandler(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)
	evs, err := getRecentEvents(limit)
	if err != nil {
		http.Error(w, "failed to load events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, evs)
}

// GET /api/events/by-device-type?device=FTD
func eventsByDeviceTypeHandler(w http.ResponseWriter, r *http.Request) {
	device := r.URL.Query().Get("device")
	if device == "" {
		http.Error(w, "missing device", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r)
	evs, err := getEventsByIndex(bucketIdxDeviceType, device, limit)
	if err != nil {
		http.Error(w, "failed to load events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, evs)
}

// GET /api/events/by-device-id?device_id=FTD-1
func eventsByDeviceIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("device_id")
	if id == "" {
		http.Error(w, "missing device_id", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r)
	evs, err := getEventsByIndex(bucketIdxDeviceID, id, limit)
	if err != nil {
		http.Error(w, "failed to load events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, evs)
}

// GET /api/events/by-severity?severity=critical
func eventsBySeverityHandler(w http.ResponseWriter, r *http.Request) {
	sev := r.URL.Query().Get("severity")
	if sev == "" {
		http.Error(w, "missing severity", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r)
	evs, err := getEventsByIndex(bucketIdxSeverity, sev, limit)
	if err != nil {
		http.Error(w, "failed to load events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, evs)
}

// GET /api/events/by-tenant?tenant=Tenant-A
func eventsByTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenant := r.URL.Query().Get("tenant")
	if tenant == "" {
		http.Error(w, "missing tenant", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r)
	evs, err := getEventsByIndex(bucketIdxTenant, tenant, limit)
	if err != nil {
		http.Error(w, "failed to load events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, evs)
}
