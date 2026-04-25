# =============================================================================
#  Digital-Memory  —  Local Development Startup Script (Windows PowerShell)
#  Run this from the Digital-Memory/ project root:
#      .\run_local.ps1
# =============================================================================

$root = $PSScriptRoot   # absolute path to the folder this script lives in

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  Digital-Memory  —  Local Dev Startup" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# --------------------------------------------------------------------------
# 1. Verify PostgreSQL is reachable before starting services
# --------------------------------------------------------------------------
Write-Host "[check] Verifying PostgreSQL is running..." -ForegroundColor Yellow
try {
    $pg = & psql -U postgres -d digital_memory -c "SELECT 1;" 2>&1
    if ($LASTEXITCODE -ne 0) { throw "psql returned non-zero exit code" }
    Write-Host "[check] PostgreSQL OK" -ForegroundColor Green
} catch {
    Write-Host ""
    Write-Host "[ERROR] Cannot connect to PostgreSQL." -ForegroundColor Red
    Write-Host "        Make sure PostgreSQL is running and the database exists:" -ForegroundColor Red
    Write-Host "        psql -U postgres -c ""CREATE DATABASE digital_memory;""" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

# --------------------------------------------------------------------------
# 2. Start ingestion-service  (Go, port 8001)
# --------------------------------------------------------------------------
Write-Host "[start] Starting ingestion-service on :8001 ..." -ForegroundColor Cyan
$ingestionDir = Join-Path $root "backend\ingestion-service"
Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "cd '$ingestionDir'; Write-Host 'ingestion-service' -ForegroundColor Cyan; go run ./cmd/main.go"
) -WindowStyle Normal

# --------------------------------------------------------------------------
# 3. Start api-service  (Go, port 8000)
# --------------------------------------------------------------------------
Write-Host "[start] Starting api-service on :8000 ..." -ForegroundColor Cyan
$apiDir = Join-Path $root "backend\api-service"
Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "cd '$apiDir'; Write-Host 'api-service' -ForegroundColor Cyan; go run ./cmd/main.go"
) -WindowStyle Normal

# --------------------------------------------------------------------------
# 4. Start ai-service  (Python / FastAPI, port 8002)
# --------------------------------------------------------------------------
Write-Host "[start] Starting ai-service on :8002 ..." -ForegroundColor Cyan
$aiDir = Join-Path $root "backend\ai-service"
Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "cd '$aiDir'; Write-Host 'ai-service' -ForegroundColor Cyan; .\.venv\Scripts\Activate.ps1; uvicorn app.main:app --host 0.0.0.0 --port 8002 --reload"
) -WindowStyle Normal

Write-Host ""
Write-Host "================================================" -ForegroundColor Green
Write-Host "  All services starting in separate windows."  -ForegroundColor Green
Write-Host ""
Write-Host "  Endpoints:"
Write-Host "    API service       -> http://localhost:8000/health"
Write-Host "    Ingestion service -> http://localhost:8001/health"
Write-Host "    AI service        -> http://localhost:8002/health"
Write-Host "================================================" -ForegroundColor Green
Write-Host ""
