$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$EnvFile = Join-Path $Root ".env"
if (Test-Path $EnvFile) {
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^\s*([^#=]+)=(.*)$') {
            [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), "Process")
        }
    }
}
$DbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "127.0.0.1" }
$DbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "3306" }
$DbUser = if ($env:DB_USER) { $env:DB_USER } else { "root" }
$DbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "" }
$DbName = if ($env:DB_NAME) { $env:DB_NAME } else { "raven_webmarket" }
$SqlFile = Join-Path $Root "database\migrations\001_init.sql"
$CreateDb = "CREATE DATABASE IF NOT EXISTS ``$DbName`` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
if ($DbPassword) {
    & mysql -h $DbHost -P $DbPort -u $DbUser -p"$DbPassword" -e $CreateDb
    Get-Content $SqlFile -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser -p"$DbPassword" $DbName
    $SqlFile2 = Join-Path $Root "database\migrations\002_admin_rbac.sql"
    Get-Content $SqlFile2 -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser -p"$DbPassword" $DbName
    $SqlFile3 = Join-Path $Root "database\migrations\003_cms_content.sql"
    Get-Content $SqlFile3 -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser -p"$DbPassword" $DbName
} else {
    & mysql -h $DbHost -P $DbPort -u $DbUser -e $CreateDb
    Get-Content $SqlFile -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser $DbName
    $SqlFile2 = Join-Path $Root "database\migrations\002_admin_rbac.sql"
    Get-Content $SqlFile2 -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser $DbName
    $SqlFile3 = Join-Path $Root "database\migrations\003_cms_content.sql"
    Get-Content $SqlFile3 -Raw | & mysql -h $DbHost -P $DbPort -u $DbUser $DbName
}
Write-Host "Migration complete: $DbName"
