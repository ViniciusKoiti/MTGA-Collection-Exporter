# Clean-VM smoke test for OpenSpec introduce-agentic-go-companion task
# 7.4. Run INSIDE a fresh Windows VM, from the kit folder:
#   powershell -ExecutionPolicy Bypass -File .\vm_smoke.ps1
# For the offline-startup leg, disconnect the VM network first and pass
# -Offline. The script writes vm-smoke-result.txt next to itself — bring
# that file back as the task evidence.
param(
    [switch]$Offline
)
$ErrorActionPreference = 'Continue'
$kit = $PSScriptRoot
$report = Join-Path $kit 'vm-smoke-result.txt'
$lines = [System.Collections.Generic.List[string]]::new()
function Note([string]$text) { $lines.Add($text); Write-Output $text }

$os = Get-CimInstance Win32_OperatingSystem
Note "machine: $($os.Caption) build $($os.BuildNumber) at $(Get-Date -Format o)"
Note "offline flag: $Offline"

# Clean-machine evidence: no Go toolchain on this VM.
$go = Get-Command go -ErrorAction SilentlyContinue
Note "go toolchain present: $([bool]$go) (expected False on a clean VM)"

# WebView2 runtime detection (Evergreen registry keys, machine + user).
$wvKeys = @(
    'HKLM:\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}',
    'HKCU:\SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}')
$wv = $wvKeys | Where-Object { Test-Path $_ } |
    ForEach-Object { (Get-ItemProperty $_).pv } | Select-Object -First 1
Note "webview2 runtime: $(if ($wv) { $wv } else { 'ABSENT (Evergreen bootstrap needed)' })"

# Packaging proof: the kit ships no dev MCP, and desktop.exe carries no
# devmcp code (the dev server lives in internal/devmcp, never imported
# by the desktop).
$extra = Get-ChildItem $kit -File | Where-Object { $_.Name -notin 'desktop.exe', 'vm_smoke.ps1', 'vm-smoke-result.txt' }
Note "unexpected kit files: $(if ($extra) { $extra.Name -join ', ' } else { 'none' })"
$bytes = [IO.File]::ReadAllBytes((Join-Path $kit 'desktop.exe'))
$text = [Text.Encoding]::ASCII.GetString($bytes)
$hit = $text.IndexOf('internal/devmcp') -ge 0
Note "devmcp marker in desktop.exe: $hit (expected False)"

# First-run startup: launch, wait, confirm the process survives and the
# local database is created under %APPDATA%\MTGA-Companion.
$dataDir = Join-Path $env:APPDATA 'MTGA-Companion'
if (Test-Path $dataDir) { Note "warning: $dataDir already existed before the run" }
$proc = Start-Process (Join-Path $kit 'desktop.exe') -PassThru
Start-Sleep -Seconds 12
$alive = -not $proc.HasExited
Note "process alive after 12s: $alive"
Note "data dir created: $(Test-Path $dataDir)"
Note "companion.db created: $(Test-Path (Join-Path $dataDir 'companion.db'))"
$remote = Get-NetTCPConnection -OwningProcess $proc.Id -ErrorAction SilentlyContinue |
    Where-Object { $_.RemoteAddress -notin '127.0.0.1', '::1', '0.0.0.0', '::' }
Note "remote connections: $(if ($remote) { ($remote | ForEach-Object { $_.RemoteAddress + ':' + $_.RemotePort }) -join ', ' } else { 'none' })"
if ($alive) { Stop-Process -Id $proc.Id -Force }

$lines | Set-Content $report
Write-Output "report written: $report"
