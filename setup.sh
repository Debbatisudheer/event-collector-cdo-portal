#!/bin/bash

root="eventcollector"

# Create main directories
mkdir -p $root/collector
mkdir -p $root/devices

# Create files
touch $root/go.mod

collector_files=(
  "main.go"
  "models.go"
  "auth.go"
  "api.go"
  "worker.go"
  "storage.go"
  "db.go"
  "metrics.go"
  "s3_uploader.go"
  "alerts.go"
)

device_files=(
  "main.go"
  "common.go"
  "ftd.go"
  "wsa.go"
  "esa.go"
  "ise.go"
  "telemetry_ws.go"
)

for file in "${collector_files[@]}"; do
  touch "$root/collector/$file"
done

for file in "${device_files[@]}"; do
  touch "$root/devices/$file"
done

echo "📁 Folder structure + files created!"


#run this command in shell
#chmod +x setup.sh
#./setup.sh
