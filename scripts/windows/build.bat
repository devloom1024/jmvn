@echo off
setlocal

set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"

if "%VERSION%"=="" set "VERSION=1.0.0"

echo Building jmvn %VERSION% ...
if not exist bin mkdir bin
go build -ldflags "-s -w -X main.version=%VERSION%" -o bin\jmvn.exe .
echo Done: bin\jmvn.exe
