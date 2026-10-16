@echo off
chcp 65001 >nul
cd /d "%~dp0"
netstat -ano | findstr ":8080.*LISTENING" >nul
if %errorlevel%==0 (
  echo StudyPilot 已在运行，直接打开浏览器...
) else (
  echo 正在启动 StudyPilot...
  start "StudyPilot" /min "%~dp0bin\studypilot.exe"
)
timeout /t 1 /nobreak >nul
start http://localhost:8080