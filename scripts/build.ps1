# Build script for Windows
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

Write-Host "Building frontend..."
Set-Location frontend
if (Test-Path package-lock.json) { npm ci } else { npm install }
npm run build
Set-Location ..

Write-Host "Building backend..."
Set-Location backend
go build -o ..\server.exe .\cmd\server
Set-Location ..

Write-Host "Building Docker image..."
docker build -t xxjz-go-web:latest .

Write-Host "Done. Image: xxjz-go-web:latest"
