#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Write-Host "https://localhost:8081"
Start-Process -Wait -FilePath "caddy" -ArgumentList "-conf", "Caddyfile-test"
