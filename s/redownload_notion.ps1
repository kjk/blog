#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

$exe = "./blog_app.exe"
go build -o $exe
exitIfFailed

Start-Process -Wait -NoNewWindow -FilePath $exe -ArgumentList '-redownload-notion'

Remove-Item -Path $exe
