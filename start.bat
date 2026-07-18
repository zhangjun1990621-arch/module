@echo off
REM ============================================================
REM  IoT Platform - Master Start Script
REM  Starts both backend (Go:8090) and frontend (Vite:3001)
REM ============================================================
title IoT Platform Launcher

cd /d "%~dp0"

echo ============================================================
echo  IoT Platform - Starting Services
echo  Backend  : http://localhost:8090
echo  Frontend : http://localhost:3001
echo  Database : PostgreSQL @ 127.0.0.1:5435
echo ============================================================
echo.

REM --- Step 1: Kill old processes on ports 8090 and 3001 ---
echo [1/3] Cleaning up old processes...

REM Kill processes on port 8090 (backend)
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8090 " ^| findstr "LISTENING"') do (
    echo   Killing backend process PID %%a
    taskkill /F /PID %%a >nul 2>&1
)

REM Kill processes on port 3001 (frontend)
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001 " ^| findstr "LISTENING"') do (
    echo   Killing frontend process PID %%a
    taskkill /F /PID %%a >nul 2>&1
)

REM Also kill any residual node.exe from Vite on port 3000
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000 " ^| findstr "LISTENING"') do (
    echo   Killing residual process on port 3000 PID %%a
    taskkill /F /PID %%a >nul 2>&1
)

echo   Cleanup done.
echo.

REM --- Step 2: Launch backend in a new window ---
echo [2/3] Launching backend service...
start "IoT-Backend (Go:8090)" cmd /k "%~dp0start-backend.bat"

REM --- Step 3: Wait for backend to be ready, then launch frontend ---
echo [3/3] Waiting for backend to be ready...
set BACKEND_READY=0
set WAIT_COUNT=0

:WAIT_LOOP
timeout /t 2 /nobreak >nul
set /a WAIT_COUNT+=1

REM Check if port 8090 is listening
netstat -ano | findstr ":8090 " | findstr "LISTENING" >nul 2>&1
if %ERRORLEVEL%==0 (
    set BACKEND_READY=1
    echo   Backend is ready! (took %WAIT_COUNT% retries)
    goto START_FRONTEND
)

if %WAIT_COUNT% GEQ 30 (
    echo   WARNING: Backend not ready after 60 seconds.
    echo   Starting frontend anyway...
    goto START_FRONTEND
)

echo   Waiting for backend... (attempt %WAIT_COUNT%/30)
goto WAIT_LOOP

:START_FRONTEND
echo.
echo Launching frontend service...
start "IoT-Frontend (Vite:3001)" cmd /k "%~dp0start-frontend.bat"

echo.
echo ============================================================
echo  All services launched!
echo    Backend  : http://localhost:8090
echo    Frontend : http://localhost:3001
echo    Login    : admin / admin123
echo ============================================================
echo.
echo  Close this window or press any key to exit.
echo  Backend and Frontend windows will keep running.
pause >nul
