@echo off
setlocal

if "%~1"=="" goto :usage

set "SCRIPTS=%~dp0scripts\windows"

if "%~1"=="build"     call "%SCRIPTS%\build.bat" %2 %3 %4 %5 && goto :end
if "%~1"=="build-all" call "%SCRIPTS%\build-all.bat" %2 %3 %4 %5 && goto :end
if "%~1"=="test"      call "%SCRIPTS%\test.bat" %2 %3 %4 %5 && goto :end
if "%~1"=="clean"     call "%SCRIPTS%\clean.bat" %2 %3 %4 %5 && goto :end
if "%~1"=="install"   call "%SCRIPTS%\install.bat" %2 %3 %4 %5 && goto :end

:usage
echo Usage: run.bat ^<command^>
echo.
echo Commands:
echo   build      Build for current platform
echo   build-all  Cross-compile for all platforms
echo   test       Run tests
echo   clean      Remove build artifacts
echo   install    Build and install to INSTALL_DIR in .env
exit /b 1

:end
endlocal
