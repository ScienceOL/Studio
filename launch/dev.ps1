# =============================================
# 构建开发环境容器服务控制脚本（PowerShell 版本）
# =============================================
param (
    [switch]$d,
    [switch]$h
)
# -------------------------------
# 全局配置
# -------------------------------
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$ProjectDir = Split-Path -Parent $ScriptDir
$EnvFile = Join-Path $ProjectDir "docker\.env.dev"

# -------------------------------
# 帮助信息
# -------------------------------
function Show-Help {
    Write-Host "`n使用说明：" -ForegroundColor Green
    Write-Host "  dev.ps1 [选项]"
    Write-Host "`n选项说明：" -ForegroundColor Green
    Write-Host "  -h     显示帮助信息" -ForegroundColor Yellow
    Write-Host "  -d     以守护进程模式启动容器（后台运行）" -ForegroundColor Yellow
    Write-Host "`n示例：" -ForegroundColor Green
    Write-Host "  ./dev.ps1"
    Write-Host "  ./dev.ps1 -d"
    Write-Host "  ./dev.ps1 -h`n"
    exit
}

# -------------------------------
# 参数解析
# -------------------------------
$BackgroundMode = $false

param (
    [switch]$d,
    [switch]$h
)

if ($h) {
    Show-Help
}

if ($d) {
    $BackgroundMode = $true
}

# -------------------------------
# 环境预检
# -------------------------------
Write-Host "`n?  检查 Docker 环境..." -ForegroundColor Magenta
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "? 错误：未检测到 Docker 安装" -ForegroundColor Red
    exit 1
}
Write-Host "? Docker 已就绪" -ForegroundColor Green

# -------------------------------
# 启动服务
# -------------------------------
Write-Host "`n? 启动开发容器服务..." -ForegroundColor Cyan
$ComposeFiles = @(
    "-f", "$ProjectDir\docker\docker-compose.base.yaml",
    "-f", "$ProjectDir\docker\docker-compose.dev.yaml",
    "--env-file", "$EnvFile"
)

if ($BackgroundMode) {
    Write-Host "? 以守护进程模式启动" -ForegroundColor Yellow
    docker compose @ComposeFiles up -d
} else {
    Write-Host "? 以前台模式启动" -ForegroundColor Yellow
    docker compose @ComposeFiles up
}
