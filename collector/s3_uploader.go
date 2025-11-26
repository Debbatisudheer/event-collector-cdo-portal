package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	telemetryBatch   []Event
	telemetryMu      sync.Mutex
	maxBatchSize     = 50
	maxBatchInterval = 15 * time.Second

	s3Client   *s3.Client
	s3Bucket   = "your-telemetry-bucket-name" // TODO: change
	s3Prefix   = "telemetry/"
	s3InitOnce sync.Once
)

func initS3() {
	s3InitOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			fmt.Println("Failed to load AWS config:", err)
			return
		}
		s3Client = s3.NewFromConfig(cfg)
		go telemetryFlushLoop()
	})
}

func addTelemetryToBatch(e Event) {
	telemetryMu.Lock()
	telemetryBatch = append(telemetryBatch, e)
	shouldFlush := len(telemetryBatch) >= maxBatchSize
	telemetryMu.Unlock()

	if shouldFlush {
		go flushTelemetryBatch()
	}
}

func telemetryFlushLoop() {
	for {
		time.Sleep(maxBatchInterval)
		flushTelemetryBatch()
	}
}

func flushTelemetryBatch() {
	telemetryMu.Lock()
	batch := telemetryBatch
	telemetryBatch = nil
	telemetryMu.Unlock()

	if len(batch) == 0 {
		return
	}

	data, _ := json.Marshal(batch)

	if s3Client == nil {
		// For local dev: just log
		fmt.Printf("📦 Would upload %d telemetry events to S3 (client not configured)\n", len(batch))
		return
	}

	key := fmt.Sprintf("%s%s.json", s3Prefix, time.Now().Format("20060102-150405"))

	_, err := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		fmt.Println("Failed to upload telemetry batch to S3:", err)
		return
	}

	fmt.Printf("✅ Uploaded %d telemetry events to s3://%s/%s\n", len(batch), s3Bucket, key)
}
