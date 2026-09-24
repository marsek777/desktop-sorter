@echo off
rem Build .exe on Windows (requires Go 1.20+: https://go.dev/dl/)
if not exist dist mkdir dist
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o dist\DesktopSorter.exe .
set GOARCH=386
go build -trimpath -ldflags "-s -w" -o dist\DesktopSorter-32bit.exe .
echo Done: dist\
pause
