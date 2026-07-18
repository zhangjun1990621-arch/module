@echo off
REM ============================================================
REM  IoT Platform - Frontend Start Script (Vite, port 3001)
REM ============================================================
title IoT-Frontend (Vite:3001)

cd /d "%~dp0frontend"

echo ============================================================
echo  Frontend Service Starting...
echo  Port: 3001
echo  API Proxy: /api -> http://localhost:8090
echo ============================================================
echo.

REM --- Step 1: Check node_modules ---
echo [1/2] Checking dependencies...
if not exist "node_modules" (
    echo   node_modules not found. Running npm install...
    call npm install
    if %ERRORLEVEL% NEQ 0 (
        echo   ERROR: npm install failed.
        pause
        exit /b 1
    )
)
echo   Dependencies OK.
echo.

REM --- Step 2: Start Vite dev server on port 3001 ---
echo [2/2] Starting Vite dev server...
echo   URL: http://localhost:3001
echo   Press Ctrl+C to stop the frontend.
echo.

npx vite --port 3001 --host 0.0.0.0

echo.
echo ============================================================
echo  Frontend service stopped.
echo ============================================================
pause
