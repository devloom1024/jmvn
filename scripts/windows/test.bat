@echo off
setlocal
set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"
go test ./...
