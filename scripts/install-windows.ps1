#Requires -RunAsAdministrator
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "C:\RavenWebmarket" }
$ServiceName = "RavenWebmarketAPI"
$FrontendServiceName = "RavenWebmarketFrontend"
$Nssm = "$InstallDir\tools\nssm.exe"

Write-Host "=== Raven Webmarket — Windows Production Install ==="
Write-Host "Install directory: $InstallDir"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\tools" | Out-Null

if (-not (Test-Path $Nssm)) {
    Write-Host "Downloading NSSM..."
    $NssmZip = "$env:TEMP\nssm.zip"
    Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile $NssmZip
    Expand-Archive -Path $NssmZip -DestinationPath "$env:TEMP\nssm" -Force
    Copy-Item "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" $Nssm
}

robocopy $Root $InstallDir /MIR /XD node_modules .git /NFL /NDL /NJH /NJS | Out-Null

$EnvPath = Join-Path $InstallDir ".env"
if (-not (Test-Path $EnvPath)) {
    Copy-Item (Join-Path $InstallDir ".env.example") $EnvPath
    Write-Host "Created .env — edit before starting services."
}

Write-Host "Building API..."
Set-Location (Join-Path $InstallDir "backend")
go build -o raven-api.exe ./cmd/server

Write-Host "Building frontend..."
Set-Location (Join-Path $InstallDir "frontend")
npm ci
npm run build

$ApiExe = Join-Path $InstallDir "backend\raven-api.exe"
$NodeExe = (Get-Command node).Source
$FrontendServer = Join-Path $InstallDir "frontend\.next\standalone\server.js"
if (-not (Test-Path $FrontendServer)) {
    $FrontendServer = Join-Path $InstallDir "frontend\server.js"
}

function Install-Service($Name, $Exe, $Args, $WorkDir) {
    $existing = Get-Service -Name $Name -ErrorAction SilentlyContinue
    if ($existing) {
        & $Nssm stop $Name 2>$null
        & $Nssm remove $Name confirm 2>$null
    }
    & $Nssm install $Name $Exe $Args
    & $Nssm set $Name AppDirectory $WorkDir
    & $Nssm set $Name AppStdout (Join-Path $InstallDir "logs\$Name.log")
    & $Nssm set $Name AppStderr (Join-Path $InstallDir "logs\$Name-error.log")
    & $Nssm set $Name AppRotateFiles 1
    & $Nssm set $Name Start SERVICE_AUTO_START
    & $Nssm set $Name AppEnvironmentExtra "NODE_ENV=production" "PORT=3000"
}

New-Item -ItemType Directory -Force -Path (Join-Path $InstallDir "logs") | Out-Null

Install-Service $ServiceName $ApiExe "" (Join-Path $InstallDir "backend")
Install-Service $FrontendServiceName $NodeExe $FrontendServer (Split-Path $FrontendServer)

Write-Host "Starting services..."
& $Nssm start $ServiceName
& $Nssm start $FrontendServiceName

Write-Host ""
Write-Host "Done. Services run hidden in background (no CMD window)."
Write-Host "  Manage: services.msc or 'nssm status $ServiceName'"
Write-Host "  Logs:   $InstallDir\logs\"
Write-Host "  Edit:   $EnvPath"
Write-Host "  Migrate: .\scripts\migrate.ps1 (from $InstallDir)"
