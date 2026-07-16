param(
  [string]$Listen = "0.0.0.0:5173",
  [string]$Backend = "http://127.0.0.1:18080"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if (-not (Test-Path ".\logs")) {
  New-Item -ItemType Directory -Path ".\logs" | Out-Null
}
if (-not (Test-Path ".\web\index.html")) {
  throw "Frontend files not found: $PSScriptRoot\web"
}
if (-not (Test-Path ".\erp-frontend.exe")) {
  throw "Frontend executable not found: $PSScriptRoot\erp-frontend.exe"
}

$args = @("-listen", $Listen, "-web", ".\web", "-backend", $Backend)
$process = Start-Process -FilePath ".\erp-frontend.exe" `
  -ArgumentList $args `
  -RedirectStandardOutput ".\logs\frontend.out.log" `
  -RedirectStandardError ".\logs\frontend.err.log" `
  -WindowStyle Hidden `
  -PassThru

Write-Host "ERP frontend started, pid=$($process.Id)"
Write-Host "Open: http://localhost:5173/login"
Write-Host "LAN access: http://SERVER_IP:5173/login"
Write-Host "Proxy /api to $Backend"
