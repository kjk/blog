#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

$exe = "./blog_app.exe" # name that works both on unix and win
go build -o blog_app.exe
exitIfFailed
./blog_app.exe

netlify deploy --dir=netlify_static --site=a1bb4018-531d-4de8-934d-8d5602bacbfb --open
Remove-Item -Path $exe
