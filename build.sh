#!/bin/bash

echo "Building React App."
cd server/app
npm install
npm run build
cd ../../cmd/openbooks

echo "Building binaries for various platforms."
env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../../build/openbooks.exe
env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ../../build/openbooks_mac
env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ../../build/openbooks_mac_arm
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../build/openbooks_linux
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ../../build/openbooks_linux_arm

echo "Building local Mac binary for immediate use."
cd ../../
go build -o build/openbooks_mac ./cmd/openbooks
go build -o build/openbooks_linux ./cmd/openbooks
go build -o build/openbooks_linux_arm ./cmd/openbooks
go build -o build/openbooks_mac_arm ./cmd/openbooks
go build -o build/openbooks.exe ./cmd/openbooks

echo "Starting OpenBooks server..."
./build/openbooks_mac server --name "" --persist --no-browser-downloads