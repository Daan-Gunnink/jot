@echo off
REM Get version from argument or use default
set VERSION=%1
if "%VERSION%"=="" set VERSION=0.1.0
echo Building Windows installer with version %VERSION%

REM Build with ldflags to inject version and create NSIS installer
wails build -platform windows/amd64 -nsis -ldflags "-X 'main.Version=%VERSION%'"

echo Windows installer created at build\bin\jot-amd64-installer.exe 