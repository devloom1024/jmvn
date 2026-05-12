@echo off
setlocal

set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"

if "%VERSION%"=="" set "VERSION=1.0.0"

echo Building all platforms jmvn %VERSION% ...

if exist bin rmdir /s /q bin
mkdir bin

set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o bin\jmvn-windows-amd64.exe .
echo   -^> bin\jmvn-windows-amd64.exe

set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o bin\jmvn-linux-amd64 .
echo   -^> bin\jmvn-linux-amd64

set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o bin\jmvn-darwin-amd64 .
echo   -^> bin\jmvn-darwin-amd64

set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w -X main.version=%VERSION%" -o bin\jmvn-darwin-arm64 .
echo   -^> bin\jmvn-darwin-arm64

echo Done: all platforms built
