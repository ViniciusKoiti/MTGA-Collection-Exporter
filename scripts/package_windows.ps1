param(
    [string]$AppName = "MTGA Exporter",
    [string]$PackageName = "MTGA-Exporter-windows",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

if ($Clean) {
    foreach ($Path in @("build", "dist", "$AppName.spec")) {
        if (Test-Path $Path) {
            Remove-Item -Recurse -Force $Path
        }
    }
}

if (-not (Get-Command pyinstaller -ErrorAction SilentlyContinue)) {
    throw "PyInstaller nao encontrado. Instale com: pip install pyinstaller"
}

pyinstaller `
    --clean `
    --onefile `
    --windowed `
    --name $AppName `
    --collect-data customtkinter `
    --hidden-import pymem `
    run_gui.py

$DistDir = Join-Path $Root "dist"
$ExePath = Join-Path $DistDir "$AppName.exe"
if (-not (Test-Path $ExePath)) {
    throw "Executavel nao encontrado em: $ExePath"
}

$PackageDir = Join-Path $DistDir $PackageName
if (Test-Path $PackageDir) {
    Remove-Item -Recurse -Force $PackageDir
}
New-Item -ItemType Directory -Path $PackageDir | Out-Null

Copy-Item -Path $ExePath -Destination $PackageDir
Copy-Item -Path (Join-Path $Root "README.md") -Destination $PackageDir

$ZipPath = Join-Path $DistDir "$PackageName.zip"
if (Test-Path $ZipPath) {
    Remove-Item -Force $ZipPath
}

Compress-Archive -Path (Join-Path $PackageDir "*") -DestinationPath $ZipPath -Force

Write-Output "Pacote criado: $ZipPath"
