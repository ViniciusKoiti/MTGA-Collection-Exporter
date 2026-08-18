# Builds the clean-VM smoke kit for OpenSpec introduce-agentic-go-companion
# task 7.4: a release desktop.exe plus the in-VM verification script,
# zipped so the whole kit can be copied into a fresh Windows VM.
# Run on the dev machine: .\companion\scripts\build_vm_kit.ps1
param(
    [string]$OutDir = 'D:\gocache\tmp\companion-vm-kit'
)
$ErrorActionPreference = 'Stop'
$repo = Resolve-Path (Join-Path $PSScriptRoot '..\..')
$desktop = Join-Path $repo 'companion\desktop'

$env:PATH = 'C:\Users\Usuario\go\bin;C:\Program Files\Go\bin;' + $env:PATH
$env:GOMODCACHE = 'D:\gocache\mod'
$env:GOCACHE = 'D:\gocache\build'
$env:GOTMPDIR = 'D:\gocache\tmp'
$env:GOPROXY = 'direct'

Push-Location $desktop
try {
    wails build -clean
    if ($LASTEXITCODE -ne 0) { throw "wails build failed ($LASTEXITCODE)" }
} finally {
    Pop-Location
}

$exe = Join-Path $desktop 'build\bin\desktop.exe'
if (-not (Test-Path $exe)) { throw "missing build output: $exe" }

if (Test-Path $OutDir) { Remove-Item -Recurse -Force $OutDir }
New-Item -ItemType Directory -Force $OutDir | Out-Null
Copy-Item $exe (Join-Path $OutDir 'desktop.exe')
Copy-Item (Join-Path $PSScriptRoot 'vm_smoke.ps1') (Join-Path $OutDir 'vm_smoke.ps1')

$zip = "$OutDir.zip"
if (Test-Path $zip) { Remove-Item -Force $zip }
Compress-Archive -Path (Join-Path $OutDir '*') -DestinationPath $zip
$size = [math]::Round((Get-Item $zip).Length / 1MB, 1)
Write-Output "kit ready: $zip (${size}MB) — copy into the VM and run vm_smoke.ps1 there"
