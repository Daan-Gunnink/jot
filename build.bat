@echo off
REM Get version from argument or use default
set VERSION=%1
if "%VERSION%"=="" set VERSION=0.1.0
echo Building with version %VERSION%

REM Build with ldflags to inject version
wails build -ldflags "-X 'main.Version=%VERSION%'"

echo Build complete with version %VERSION%
echo.
echo To create an installer, use build-installer.bat instead 