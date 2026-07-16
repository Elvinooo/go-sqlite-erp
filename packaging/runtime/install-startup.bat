@echo off
setlocal
cd /d "%~dp0"
set TASK_NAME=GoSQLiteERP
schtasks /Create /TN "%TASK_NAME%" /TR "\"%~dp0start-all.bat\"" /SC ONSTART /RU SYSTEM /RL HIGHEST /F
if errorlevel 1 (
  echo Failed to install startup task. Please right-click this file and choose Run as administrator.
  exit /b 1
)
echo Installed startup task: %TASK_NAME%
echo It will start when Windows boots.
endlocal
