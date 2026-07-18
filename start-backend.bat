@echo off
REM ============================================================
REM  IoT Platform - Backend Start Script (Go + Gin, port 8090)
REM  Checks PostgreSQL real connectivity (not just port) before launching
REM ============================================================
title IoT-Backend (Go:8090)

cd /d "%~dp0backend"

echo ============================================================
echo  Backend Service Starting...
echo  Port: 8090
echo ============================================================
echo.

REM --- Step 1: Verify Go is installed ---
echo [1/3] Checking Go installation...
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo   ERROR: Go is not installed or not in PATH.
    echo   Please install Go from https://go.dev/dl/
    pause
    exit /b 1
)
for /f "tokens=*" %%v in ('go version') do set GO_VERSION=%%v
echo   %GO_VERSION%
echo.

REM --- Step 2: Check PostgreSQL real connectivity on port 5435 ---
echo [2/3] Checking PostgreSQL on port 5435...
set PG_READY=0
set PG_RETRY=0

REM Phase 1: Wait for port to be listening
:PORT_CHECK
netstat -ano | findstr ":5435 " | findstr "LISTENING" >nul 2>&1
if %ERRORLEVEL%==0 (
    echo   Port 5435 is listening. Verifying database readiness...
    goto DB_CHECK
)

set /a PG_RETRY+=1
if %PG_RETRY% GEQ 30 (
    echo   ERROR: PostgreSQL port 5435 is not listening.
    echo   Please start PostgreSQL first.
    echo   Config: 127.0.0.1:5435 / iot_platform
    pause
    exit /b 1
)

echo   Waiting for port 5435... (attempt %PG_RETRY%/30)
timeout /t 2 /nobreak >nul
goto PORT_CHECK

REM Phase 2: Test actual database connection (not just port)
:DB_CHECK
set DB_RETRY=0

:DB_RETRY_LOOP
python -c "import psycopg2; conn=psycopg2.connect(host='127.0.0.1',port=5435,user='postgres',password='Qsh@2026#PvSecure',dbname='iot_platform',connect_timeout=3); conn.close(); exit(0)" >nul 2>&1
if %ERRORLEVEL%==0 (
    echo   PostgreSQL is ready and accepting connections!
    set PG_READY=1
    goto START_BACKEND
)

set /a DB_RETRY+=1
if %DB_RETRY% GEQ 30 (
    echo   ERROR: PostgreSQL is listening but not accepting connections after 60 seconds.
    echo   This may indicate a database issue. Please check PostgreSQL logs.
    pause
    exit /b 1
)

echo   PostgreSQL still starting up... (attempt %DB_RETRY%/30)
timeout /t 2 /nobreak >nul
goto DB_RETRY_LOOP

:START_BACKEND
echo.

REM --- Step 3: Start Go backend ---
echo [3/3] Starting Go backend...
echo   Config: config\config.yaml
echo   Press Ctrl+C to stop the backend.
echo.

go run cmd\server\main.go

echo.
echo ============================================================
echo  Backend service stopped.
echo ============================================================
pause
