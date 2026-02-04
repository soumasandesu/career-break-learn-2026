#!/bin/bash

# Generate Go code from proto files
PROTO_DIR="src/main/proto"
OUTPUT_DIR="../user-service-golang/proto"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate Go code for each proto file
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
  --proto_path="$PROTO_DIR" \
  "$PROTO_DIR"/*.proto

echo "Go proto files generated in $OUTPUT_DIR"
