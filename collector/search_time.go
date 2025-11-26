package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

func eventsTimeRangeHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "start and end required", http.StatusBadRequest)
		return
	}

	startInt, err1 := strconv.ParseInt(startStr, 10, 64)
	endInt, err2 := strconv.ParseInt(endStr, 10, 64)

	if err1 != nil || err2 != nil {
		http.Error(w, "invalid timestamps", http.StatusBadRequest)
		return
	}

	var results []Event

	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("events"))
		c := b.Cursor()

		min := []byte(strconv.FormatInt(startInt, 10))
		max := []byte(strconv.FormatInt(endInt, 10))

		for k, v := c.Seek(min); k != nil && string(k) <= string(max); k, v = c.Next() {
			var e Event
			json.Unmarshal(v, &e)
			results = append(results, e)
		}
		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
