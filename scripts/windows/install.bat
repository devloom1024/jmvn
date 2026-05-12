@echo off
setlocal enabledelayedexpansion

set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"

if not exist .env (
    echo Please copy .env.example to .env and set INSTALL_DIR
    exit /b 1
)

for /f "usebackq tokens=1,2 delims==" %%a in (".env") do (
    set "%%a=%%b"
)

if "%INSTALL_DIR%"=="" (
    echo Error: INSTALL_DIR not set in .env
    exit /b 1
)

call "%~dp0build.bat"

copy /y bin\jmvn.exe "%INSTALL_DIR%\"
echo Installed to %INSTALL_DIR%
