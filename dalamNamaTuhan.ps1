# Setup & run ChristAPI with Docker (PowerShell)
# Usage:
#   .\dalamNamaTuhan.ps1
#   .\dalamNamaTuhan.ps1 -NoBuild
#   .\dalamNamaTuhan.ps1 -NoBuild -NoMigrate
#   .\dalamNamaTuhan.ps1 -MigrateOnly

param(
    [switch]$NoBuild,
    [switch]$NoMigrate,
    [switch]$Restart,
    [switch]$MigrateOnly,
    [string]$Service = ""  # optional: service name to restart (e.g. golang-christapi)
)

$ErrorActionPreference = "Stop"

function Invoke-DockerCompose {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    $stdoutFile = [System.IO.Path]::GetTempFileName()
    $stderrFile = [System.IO.Path]::GetTempFileName()

    try {
        $process = Start-Process -FilePath "docker" -ArgumentList (@("compose") + $Arguments) -NoNewWindow -Wait -PassThru -RedirectStandardOutput $stdoutFile -RedirectStandardError $stderrFile
        if ($process.ExitCode -ne 0) {
            throw "docker compose $($Arguments -join ' ') failed with exit code $($process.ExitCode)"
        }
    } finally {
        Remove-Item $stdoutFile, $stderrFile -ErrorAction SilentlyContinue
    }
}

function Invoke-DockerRaw {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    $stdoutFile = [System.IO.Path]::GetTempFileName()
    $stderrFile = [System.IO.Path]::GetTempFileName()

    try {
        $process = Start-Process -FilePath "docker" -ArgumentList $Arguments -NoNewWindow -Wait -PassThru -RedirectStandardOutput $stdoutFile -RedirectStandardError $stderrFile
        if ($process.ExitCode -ne 0) {
            throw "docker $($Arguments -join ' ') failed with exit code $($process.ExitCode)"
        }
    } finally {
        Remove-Item $stdoutFile, $stderrFile -ErrorAction SilentlyContinue
    }
}

Write-Host "[*] Bismillah... Starting ChristAPI setup..." -ForegroundColor Cyan
Write-Host ""

# 1. Check if Docker is running
Write-Host "[*] Checking Docker..." -ForegroundColor Yellow
try {
    $null = & docker ps 2>&1
    Write-Host "[OK] Docker is running" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Docker is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}
Write-Host ""

# 2. Check if .env.docker exists
Write-Host "[*] Checking Docker environment variables..." -ForegroundColor Yellow
if (-not (Test-Path ".env.docker")) {
    Write-Host "[ERROR] .env.docker not found" -ForegroundColor Red
    exit 1
} else {
    Write-Host "[OK] .env.docker already exists" -ForegroundColor Green
}
Write-Host ""

# 3. Build Docker image
if ($MigrateOnly) {
	Write-Host "[*] Skipping image build (-MigrateOnly)" -ForegroundColor Yellow
} elseif ($NoBuild -or $Restart) {
    Write-Host "[*] Skipping image build (-NoBuild or -Restart)" -ForegroundColor Yellow
} else {
    Write-Host "[*] Building Docker image..." -ForegroundColor Yellow
    Invoke-DockerCompose -Arguments @("build", "--no-cache")
    Write-Host "[OK] Build complete" -ForegroundColor Green
}
Write-Host ""

# 4. Start or Restart services
if ($MigrateOnly) {
    Write-Host "[*] Starting PostgreSQL only for migration..." -ForegroundColor Yellow
    Invoke-DockerCompose -Arguments @("up", "-d", "postgres")
    Write-Host "[OK] PostgreSQL started" -ForegroundColor Green
    Write-Host ""
} elseif ($Restart) {
    Write-Host "[*] Restart mode (-Restart)" -ForegroundColor Yellow
    if ([string]::IsNullOrEmpty($Service)) {
        Write-Host "[*] Restarting all services..." -ForegroundColor Yellow
        Invoke-DockerCompose -Arguments @("restart")
    } else {
        Write-Host "[*] Restarting service: $Service" -ForegroundColor Yellow
        try {
            Invoke-DockerCompose -Arguments @("restart", $Service)
        } catch {
            Write-Host "[WARN] 'docker compose restart $Service' failed, trying 'docker restart $Service' (container-level)" -ForegroundColor Yellow
            try {
                Invoke-DockerRaw -Arguments @("restart", $Service)
            } catch {
                Write-Host "[ERROR] Failed to restart service/container: $Service" -ForegroundColor Red
                throw
            }
        }
    }

    Write-Host "[OK] Services restarted" -ForegroundColor Green
    Write-Host ""

    # If we're in restart mode, skip build and migrations steps (fast-path)
    Write-Host "[*] Skipping build/migrate in restart mode" -ForegroundColor Yellow
} else {
    Write-Host "[*] Starting services (postgres, api)..." -ForegroundColor Yellow
    Invoke-DockerCompose -Arguments @("down")
    Invoke-DockerCompose -Arguments @("up", "-d")
    Write-Host "[OK] Services started" -ForegroundColor Green
    Write-Host ""
}

# 5. Wait for postgres to be healthy (simplified approach)
Write-Host "[*] Waiting for PostgreSQL to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 3  # Give PostgreSQL time to start
for ($attempt = 1; $attempt -le 15; $attempt++) {
    $result = & docker exec postgre-chrisapi pg_isready -U christ_user 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] PostgreSQL is healthy" -ForegroundColor Green
        break
    }
    Write-Host "    Waiting... (attempt $attempt/15)"
    Start-Sleep -Seconds 2
}
Write-Host ""

# 6. Run migrations
if ($NoMigrate -and -not $MigrateOnly) {
    Write-Host "[*] Skipping migrations (-NoMigrate)" -ForegroundColor Yellow
} else {
    Write-Host "[*] Running database migrations..." -ForegroundColor Yellow
    Invoke-DockerCompose -Arguments @("run", "--rm", "migrate", "-path=/migrations", "-database", "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable", "up")
    Write-Host "[OK] Migrations complete" -ForegroundColor Green
}
Write-Host ""

if ($MigrateOnly) {
	Write-Host "[*] Migration-only mode selesai." -ForegroundColor Yellow
	Write-Host "[SUCCESS] Migration applied successfully." -ForegroundColor Green
	exit 0
}

# 7. Show status
Write-Host "[*] Service status:" -ForegroundColor Yellow
Invoke-DockerCompose -Arguments @("ps")
Write-Host ""

# 8. Show access info
Write-Host "========================================" -ForegroundColor Green
Write-Host "[SUCCESS] ChristAPI is ready!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "[*] API Server:" -ForegroundColor Cyan
Write-Host "    http://localhost:3001" -ForegroundColor White
Write-Host ""
Write-Host "[*] Database:" -ForegroundColor Cyan
Write-Host "    Host: localhost" -ForegroundColor White
Write-Host "    Port: 5433" -ForegroundColor White
Write-Host "    Database: christ_db" -ForegroundColor White
Write-Host "    User: christ_user" -ForegroundColor White
Write-Host "    Password: christ_password" -ForegroundColor White
Write-Host ""
Write-Host "[*] Useful commands:" -ForegroundColor Cyan
Write-Host "    docker compose logs -f                 # View logs" -ForegroundColor White
Write-Host "    docker compose exec golang-christapi sh # Access API container" -ForegroundColor White
Write-Host "    docker compose down                    # Stop services" -ForegroundColor White
Write-Host ""
Write-Host "[*] DBeaver connection string:" -ForegroundColor Cyan
Write-Host "    postgres://christ_user:christ_password@localhost:5433/christ_db" -ForegroundColor White
Write-Host ""
