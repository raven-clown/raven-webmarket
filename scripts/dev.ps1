$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root
$EnvPath = Join-Path $Root ".env"
if (-not (Test-Path $EnvPath)) {
    Copy-Item (Join-Path $Root ".env.example") $EnvPath
    Write-Host "Created .env from .env.example — edit before production use."
}
docker compose -f deploy/docker-compose.yml up -d mariadb redis minio
& (Join-Path $Root "scripts\migrate.ps1")
$BackendJob = Start-Job -ScriptBlock {
    param($R)
    Set-Location (Join-Path $R "backend")
    go run ./cmd/server
} -ArgumentList $Root
$FrontendJob = Start-Job -ScriptBlock {
    param($R)
    Set-Location (Join-Path $R "frontend")
    npm run dev
} -ArgumentList $Root
Write-Host "Backend job: $($BackendJob.Id) | Frontend job: $($FrontendJob.Id)"
Write-Host "Press Ctrl+C to stop."
try {
    while ($true) { Start-Sleep -Seconds 2 }
} finally {
    Stop-Job $BackendJob, $FrontendJob -ErrorAction SilentlyContinue
    Remove-Job $BackendJob, $FrontendJob -Force -ErrorAction SilentlyContinue
}
