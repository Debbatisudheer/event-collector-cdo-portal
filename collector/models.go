package main

type Event struct {
	DeviceID string `json:"device_id"`
	Device   string `json:"device"`
	Type     string `json:"type"`
	Msg      string `json:"msg"`
	Details  string `json:"details"`
	Severity string `json:"severity"`
	Time     string `json:"time"`
	Tenant   string `json:"tenant,omitempty"`
}

