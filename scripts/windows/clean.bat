@echo off
setlocal
set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"
if exist bin rmdir /s /q bin
echo Cleaned bin/
