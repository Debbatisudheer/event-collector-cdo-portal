package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

var boltDB *bolt.DB

const (
	bucketEvents        = "events"
	bucketIdxDeviceType = "idx_device_type"
	bucketIdxDeviceID   = "idx_device_id"
	bucketIdxSeverity   = "idx_severity"
	bucketIdxTenant     = "idx_tenant"
)

// initDB is called from main.go
func initDB() {
	var err error
	boltDB, err = bolt.Open("events.db", 0600, nil)
	if err != nil {
		log.Fatal("failed to open BoltDB:", err)
	}

	err = boltDB.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			bucketEvents,
			bucketIdxDeviceType,
			bucketIdxDeviceID,
			bucketIdxSeverity,
			bucketIdxTenant,
		}
		for _, name := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal("failed to create buckets:", err)
	}

	fmt.Println("📦 BoltDB with SSX-style indexes ready")
}

// saveEventToDB is called from storage.go
func saveEventToDB(e Event) {
	if boltDB == nil {
		return
	}

	data, _ := json.Marshal(e)
	key := []byte(fmt.Sprintf("%d", time.Now().UnixNano()))

	err := boltDB.Update(func(tx *bolt.Tx) error {
		// 1) save full event
		evB := tx.Bucket([]byte(bucketEvents))
		if err := evB.Put(key, data); err != nil {
			return err
		}

		// 2) index by device type
		if e.Device != "" {
			b := tx.Bucket([]byte(bucketIdxDeviceType))
			idxKey := []byte(e.Device + "|" + string(key))
			if err := b.Put(idxKey, key); err != nil {
				return err
			}
		}

		// 3) index by device id
		if e.DeviceID != "" {
			b := tx.Bucket([]byte(bucketIdxDeviceID))
			idxKey := []byte(e.DeviceID + "|" + string(key))
			if err := b.Put(idxKey, key); err != nil {
				return err
			}
		}

		// 4) index by severity
		if e.Severity != "" {
			b := tx.Bucket([]byte(bucketIdxSeverity))
			idxKey := []byte(e.Severity + "|" + string(key))
			if err := b.Put(idxKey, key); err != nil {
				return err
			}
		}

		// 5) index by tenant (optional, for later use)
		if e.Tenant != "" {
			b := tx.Bucket([]byte(bucketIdxTenant))
			idxKey := []byte(e.Tenant + "|" + string(key))
			if err := b.Put(idxKey, key); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("❌ Failed to save event to DB:", err)
	}
}

// getRecentEvents returns latest N events (for dashboard)
func getRecentEvents(limit int) ([]Event, error) {
	if boltDB == nil {
		return nil, nil
	}
	events := make([]Event, 0, limit)

	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketEvents))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.Last(); k != nil && len(events) < limit; k, v = c.Prev() {
			var e Event
			if err := json.Unmarshal(v, &e); err == nil {
				events = append(events, e)
			}
		}
		return nil
	})

	return events, err
}

// getEventsByIndex returns events for a given index bucket + prefix
func getEventsByIndex(bucketName, prefix string, limit int) ([]Event, error) {
	if boltDB == nil {
		return nil, nil
	}
	events := make([]Event, 0, limit)

	err := boltDB.View(func(tx *bolt.Tx) error {
		idxB := tx.Bucket([]byte(bucketName))
		evB := tx.Bucket([]byte(bucketEvents))
		if idxB == nil || evB == nil {
			return nil
		}

		c := idxB.Cursor()
		prefixBytes := []byte(prefix + "|")

		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes) && len(events) < limit; k, v = c.Next() {
			evBytes := evB.Get(v)
			if evBytes == nil {
				continue
			}
			var e Event
			if err := json.Unmarshal(evBytes, &e); err == nil {
				events = append(events, e)
			}
		}
		return nil
	})

	return events, err
}
