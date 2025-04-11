@echo off
REM Get version from argument or use default
set VERSION=%1
if "%VERSION%"=="" set VERSION=0.1.0
echo Building Windows installer with version %VERSION%

REM Create tmp directory for WebView2 setup
if not exist "build\windows\installer\tmp" mkdir "build\windows\installer\tmp"

REM Download WebView2 runtime if not present
if not exist "build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe" (
    echo Downloading WebView2 runtime...
    powershell -Command "Invoke-WebRequest -Uri 'https://go.microsoft.com/fwlink/p/?LinkId=2124703' -OutFile 'build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe'"
)

REM Build with ldflags to inject version and create NSIS installer
wails build -platform windows/amd64 -nsis -ldflags "-X 'main.Version=%VERSION%'"

echo Windows installer created at build\bin\jot-amd64-installer.exe 