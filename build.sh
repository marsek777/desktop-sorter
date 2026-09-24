#!/bin/sh
# Сборка .exe (можно запускать на Linux/macOS/Windows с установленным Go 1.20+)
set -e
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/DesktopSorter.exe .
GOOS=windows GOARCH=386   CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/DesktopSorter-32bit.exe .
echo "Готово: dist/"
