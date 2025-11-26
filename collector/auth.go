package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func init() {
	// random seed for token generation
	rand.Seed(time.Now().UnixNano())
}

var (
	devicesByToken = make(map[string]Device)
	devMu          sync.RWMutex
)

// ================= DEVICE STRUCT =================

type Device struct {
	ID      string `json:"device_id"`
	Type    string `json:"type"`
	Tenant  string `json:"tenant"`
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
}

// ================ REGISTRATION REQUEST STRUCTS ===============

type RegisterRequest struct {
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`   // FTD, WSA, ESA, ISE
	Tenant   string `json:"tenant"` // optional
}

type RegisterResponse struct {
	Token string `json:"token"`
}

// ================ TOKEN GENERATION =================

func generateToken() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ================ REGISTER HANDLER =================

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token := generateToken()

	dev := Device{
		ID:      req.DeviceID,
		Type:    req.Type,
		Tenant:  req.Tenant,
		Token:   token,
		Enabled: true,
	}

	devMu.Lock()
	devicesByToken[token] = dev
	devMu.Unlock()

	resp := RegisterResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ================ TOKEN VALIDATION =================

func validateToken(r *http.Request) (*Device, bool) {
	const prefix = "Bearer "

	auth := r.Header.Get("Authorization")
	if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		return nil, false
	}

	token := auth[len(prefix):]

	devMu.RLock()
	dev, ok := devicesByToken[token]
	devMu.RUnlock()

	if !ok || !dev.Enabled {
		return nil, false
	}

	return &dev, true
}
