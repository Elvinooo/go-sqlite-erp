@echo off
cd /d "%~dp0"
if not exist logs mkdir logs
if not exist data mkdir data
if not exist data\uploads mkdir data\uploads
erp-server.exe >> "%~dp0logs\backend.out.log" 2>&1
