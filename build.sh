#!/bin/bash

# Windows x86
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o go-upload-windows-amd64.exe main.go

# Windows ARM
GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o go-upload-windows-arm64.exe main.go

# Linux x86
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o go-upload-linux-amd64 main.go

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o go-upload-darwin-arm64 main.go

echo "build completed!"
