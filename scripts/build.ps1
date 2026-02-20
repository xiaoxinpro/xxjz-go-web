# 构建脚本 (Windows)：输出到项目根目录 build/，可在该目录直接运行
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$Root = Join-Path $PSScriptRoot ".."
$BuildDir = Join-Path $Root "build"

Set-Location $Root

# 若存在则删除 build，再创建
if (Test-Path $BuildDir) {
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path $BuildDir | Out-Null

Write-Host "Building frontend..."
Set-Location (Join-Path $Root "frontend")
if (Test-Path package-lock.json) { npm ci } else { npm install }
npm run build
Set-Location $Root

Write-Host "Building backend..."
Set-Location (Join-Path $Root "backend")
$exe = if ($env:GOOS -eq "linux") { "server" } else { "server.exe" }
go build -o (Join-Path $BuildDir $exe) ./cmd/server
Set-Location $Root

# 复制运行所需文件到 build
Copy-Item (Join-Path $Root "config.yaml") -Destination $BuildDir
$staticDir = Join-Path $BuildDir "static"
New-Item -ItemType Directory -Path $staticDir -Force | Out-Null
Copy-Item -Path (Join-Path $Root "frontend\dist\*") -Destination $staticDir -Recurse -Force
Copy-Item -Path (Join-Path $Root "backend\migrations") -Destination $BuildDir -Recurse -Force

if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "Building Docker image..."
    Set-Location $Root
    docker build -t xxjz-go-web:latest .
    Write-Host "Docker image: xxjz-go-web:latest"
} else {
    Write-Host "Docker not found, skipping Docker image build."
}

Set-Location $Root
Write-Host "Build finished. Run: cd build && .\server.exe"
Write-Host "  (Linux/macOS: cd build && ./server)"
