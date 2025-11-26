package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func saveEventToFile(e Event) {
	f, err := os.OpenFile("events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("❌ Failed to open log file:", err)
		return
	}
	defer f.Close()

	data, _ := json.Marshal(e)
	f.WriteString(string(data) + "\n")
}

func saveEvent(e Event) {
	saveEventToFile(e)
	saveEventToDB(e)
}
