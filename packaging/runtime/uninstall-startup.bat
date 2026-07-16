@echo off
setlocal
set TASK_NAME=GoSQLiteERP
schtasks /Delete /TN "%TASK_NAME%" /F
if errorlevel 1 (
  echo Failed to remove startup task. Please right-click this file and choose Run as administrator.
  exit /b 1
)
echo Removed startup task: %TASK_NAME%
endlocal
