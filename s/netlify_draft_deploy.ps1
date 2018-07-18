#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

$exe = "./blog_app.exe" # name that works both on unix and win
go build -o blog_app.exe
exitIfFailed
Start-Process -Wait -FilePath $exe -ArgumentList "-deploy"
Remove-Item -Path $exe

netlifyctl deploy --draft
