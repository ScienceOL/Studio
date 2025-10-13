# =============================================
# Control script for building development environment container services (PowerShell version)
# =============================================
param (
    [switch]$d,
    [switch]$h,
    [switch]$e,
    [switch]$s
)

# -------------------------------
# Global Configuration
# -------------------------------
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$ProjectDir = Split-Path -Parent $ScriptDir
$EnvFile = Join-Path $ProjectDir "docker\.env.dev"
# Define the path for the sciol global shared virtual environment
$SciolVenvPath = Join-Path $HOME ".sciol\venv"
$PythonVersion = "3.14"

# -------------------------------
# Help Information
# -------------------------------
function Show-Help {
    Write-Host "`nUsage:" -ForegroundColor Green
    Write-Host "  dev.ps1 [options]"
    Write-Host "`nOptions:" -ForegroundColor Green
    Write-Host "  -h     Show help information" -ForegroundColor Yellow
    Write-Host "  -d     Start containers in daemon mode (run in background)" -ForegroundColor Yellow
    Write-Host "  -e     Stop and remove all containers" -ForegroundColor Yellow
    Write-Host "  -s     Quickly stop containers (without removing)" -ForegroundColor Yellow
    Write-Host "`nExamples:" -ForegroundColor Green
    Write-Host "  # Start development service in foreground"
    Write-Host "  ./dev.ps1"
    Write-Host "  # Start development service in background"
    Write-Host "  ./dev.ps1 -d"
    Write-Host "  # Quickly stop containers"
    Write-Host "  ./dev.ps1 -s"
    Write-Host "  # Stop and remove containers"
    Write-Host "  ./dev.ps1 -e"
    Write-Host "  # Show help information"
    Write-Host "  ./dev.ps1 -h`n"
    exit
}

# -------------------------------
# Development Environment Setup
# -------------------------------

# Step 1: Check Docker and .env file
function Check-Basics {
    Write-Host "`n[1/3] Checking basic environment..." -ForegroundColor Magenta
    # Docker check
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Host "[!] Error: Docker installation not detected" -ForegroundColor Red
        exit 1
    }

    # .env.dev check
    if (-not (Test-Path $EnvFile)) {
        Write-Host "   [WARN] .env.dev file not found, creating from .env.example..." -ForegroundColor Yellow
        Copy-Item -Path (Join-Path $ProjectDir "docker\.env.example") -Destination $EnvFile
    }
    Write-Host "   [OK] Docker and .env file are ready." -ForegroundColor Green
}

# Step 2: Configure Sciol virtual environment and pre-commit
function Setup-SciolEnv {
    # Check if uv is installed
    if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
        Write-Host "[!] Error: This project uses uv to manage pre-commit, but uv is not installed." -ForegroundColor Red
        Write-Host "Please see https://github.com/astral-sh/uv for installation instructions." -ForegroundColor Yellow
        exit 1
    }

    $PythonExePath = Join-Path $SciolVenvPath "Scripts\python.exe"
    $PreCommitExePath = Join-Path $SciolVenvPath "Scripts\pre-commit.exe"

    # First run detection
    if (-not (Test-Path $PythonExePath)) {
        Write-Host "`n[2/3] First-time setup for Sciol shared environment..." -ForegroundColor Magenta
        Write-Host "   -> Creating Python $PythonVersion virtual environment at $SciolVenvPath..." -ForegroundColor Cyan
        New-Item -ItemType Directory -Force -Path (Split-Path $SciolVenvPath) | Out-Null
        uv venv $SciolVenvPath --python $PythonVersion
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[!] Failed to create virtual environment." -ForegroundColor Red
            exit 1
        }

        Write-Host "   -> Installing pre-commit..." -ForegroundColor Cyan
        uv pip install --python $PythonExePath pre-commit
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[!] Failed to install pre-commit." -ForegroundColor Red
            exit 1
        }

        Write-Host "   -> Installing Git hooks..." -ForegroundColor Cyan
        Push-Location $ProjectDir
        & $PreCommitExePath install
        Pop-Location
        Write-Host "   [OK] Sciol shared environment setup complete." -ForegroundColor Green
    } else {
        Write-Host "`n[2/3] Verifying Sciol shared environment..." -ForegroundColor Magenta
        # On subsequent runs, quietly ensure pre-commit and hooks are installed
        if (-not (Test-Path $PreCommitExePath)) {
            uv pip install --python $PythonExePath pre-commit | Out-Null
        }
        Push-Location $ProjectDir
        & $PreCommitExePath install | Out-Null
        Pop-Location
        Write-Host "   [OK] Sciol environment ($SciolVenvPath) is configured." -ForegroundColor Green
    }
}

# Step 3: Configure VS Code Workspace
function Setup-VSCodeWorkspace {
    $SettingsJsonPath = Join-Path $ProjectDir ".vscode\settings.json"
    $ExtensionsJsonPath = Join-Path $ProjectDir ".vscode\extensions.json"
    $SettingsExamplePath = Join-Path $ProjectDir ".vscode\settings.example.json"
    $ExtensionsExamplePath = Join-Path $ProjectDir ".vscode\extensions.example.json"

    # First run detection
    if (-not (Test-Path $SettingsJsonPath) -or -not (Test-Path $ExtensionsJsonPath)) {
        Write-Host "`n[3/3] First-time setup for VS Code workspace..." -ForegroundColor Magenta
        New-Item -ItemType Directory -Force -Path (Join-Path $ProjectDir ".vscode") | Out-Null

        if (-not (Test-Path $SettingsJsonPath) -and (Test-Path $SettingsExamplePath)) {
            Write-Host "   -> Creating settings.json from settings.example.json..." -ForegroundColor Cyan
            Copy-Item -Path $SettingsExamplePath -Destination $SettingsJsonPath
        }

        if (-not (Test-Path $ExtensionsJsonPath) -and (Test-Path $ExtensionsExamplePath)) {
            Write-Host "   -> Creating extensions.json from extensions.example.json..." -ForegroundColor Cyan
            Copy-Item -Path $ExtensionsExamplePath -Destination $ExtensionsJsonPath
            Write-Host "   [INFO] Please install recommended extensions in VS Code." -ForegroundColor Yellow
        }
        Write-Host "   [OK] VS Code workspace setup complete." -ForegroundColor Green
    } else {
        Write-Host "`n[3/3] Verifying VS Code workspace..." -ForegroundColor Magenta
        Write-Host "   [OK] VS Code configuration files are ready." -ForegroundColor Green
    }
}

# =============================================
# Argument Parsing
# =============================================
if ($h) {
    Show-Help
}

# =============================================
# Main Execution Flow
# =============================================

# -------------------------------
# Development Environment Setup
# -------------------------------
Write-Host "`nConfiguring development environment..." -ForegroundColor Blue
Check-Basics
Setup-SciolEnv
Setup-VSCodeWorkspace

# -------------------------------
# Service Start and Management
# -------------------------------
# Development service Docker Compose arguments
$DevComposeArgs = @(
    "-f", "$ProjectDir\docker\docker-compose.base.yaml",
    "-f", "$ProjectDir\docker\docker-compose.dev.yaml",
    "--env-file", $EnvFile
)
# Middleware service Docker Compose arguments
$InfraComposeArgs = @(
    "-p", "sciol-infra",
    "-f", "$ProjectDir\docker\docker-compose.infra.yaml",
    "--env-file", $EnvFile
)

# Handle stop and remove containers command
if ($e) {
    Write-Host "`n-> Stopping and removing development service containers..." -ForegroundColor Yellow
    docker compose @DevComposeArgs down
    Write-Host "-> Stopping and removing middleware service containers..." -ForegroundColor Yellow
    docker compose @InfraComposeArgs down
    exit
}

# Handle stop containers command
if ($s) {
    Write-Host "`n-> Stopping development service containers..." -ForegroundColor Yellow
    docker compose @DevComposeArgs stop
    Write-Host "-> Stopping middleware service containers..." -ForegroundColor Yellow
    docker compose @InfraComposeArgs stop
    exit
}

# Check and start middleware services
Write-Host "`n[+] Checking infrastructure services status..." -ForegroundColor Cyan
$RunningInfraServices = docker compose @InfraComposeArgs ps --status=running -q
if ($RunningInfraServices) {
    Write-Host "[OK] Infrastructure services are already running." -ForegroundColor Green
} else {
    Write-Host "-> Infrastructure services are not running, starting in background..." -ForegroundColor Yellow
    docker compose @InfraComposeArgs up -d
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[!] Failed to start infrastructure services." -ForegroundColor Red
        exit 1
    }
    Write-Host "[OK] Infrastructure services started successfully." -ForegroundColor Green
}

# Start development services
Write-Host "`n==> Starting development container services..." -ForegroundColor Blue
if ($d) {
    Write-Host "-> Starting in daemon mode" -ForegroundColor Yellow
    docker compose @DevComposeArgs up -d
} else {
    Write-Host "-> Starting in foreground mode" -ForegroundColor Yellow
    docker compose @DevComposeArgs up
}
