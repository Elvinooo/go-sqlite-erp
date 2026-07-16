@echo off
cd /d "%~dp0"
start "ERP Backend" /min cmd /c "%~dp0start-backend.bat"
powershell -NoProfile -Command "Start-Sleep -Seconds 2" >nul
start "ERP Frontend" /min powershell -ExecutionPolicy Bypass -File "%~dp0start-frontend.ps1"
