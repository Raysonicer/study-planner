# StudyPilot 一键构建：web 前端 → server 内嵌 → 单二进制
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host '[1/3] 构建前端 ...'
Push-Location (Join-Path $root 'web')
pnpm install --frozen-lockfile
pnpm build
Pop-Location

Write-Host '[2/3] 拷贝 dist 到 server/webdist ...'
Copy-Item -Path (Join-Path $root 'web\dist\*') -Destination (Join-Path $root 'server\internal\frontend\dist') -Recurse -Force

Write-Host '[3/3] 编译 Go 二进制 ...'
New-Item -ItemType Directory -Force -Path (Join-Path $root 'bin') | Out-Null
Push-Location (Join-Path $root 'server')
go build -trimpath -ldflags '-s -w' -o (Join-Path $root 'bin\studypilot.exe') ./cmd/server
Pop-Location

Write-Host "完成：$root\bin\studypilot.exe（运行后访问 http://localhost:8080）"