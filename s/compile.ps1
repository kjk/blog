#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { Write-Host "error"; exit } }

$exe = "./blog_app.exe"
Remove-Item $exe
go build -o blog_app.exe
exitIfFailed
Remove-Item $exe
