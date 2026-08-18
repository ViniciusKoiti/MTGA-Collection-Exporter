# Release-artifact inspection (OpenSpec add-graph-workflow-harness,
# task 8.6): fail the pipeline if the shipped zip carries anything
# development-only. Runs in CI right after the build, before upload.
param(
    [Parameter(Mandatory = $true)][string]$ZipPath
)

$ErrorActionPreference = "Stop"
if (-not (Test-Path $ZipPath)) {
    throw "inspect_release: artifact not found: $ZipPath"
}

Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip = [System.IO.Compression.ZipFile]::OpenRead((Resolve-Path $ZipPath))
try {
    $entries = $zip.Entries | ForEach-Object { $_.FullName }
} finally {
    $zip.Dispose()
}
if ($entries.Count -eq 0) {
    throw "inspect_release: the artifact is empty - nothing was proven"
}

# Forbidden file-name fragments: the dev MCP, test credentials,
# private fixtures and development configuration must never ship.
$forbidden = @(
    "dev-mcp", "devmcp",
    "test_credentials", "credentials.test", ".pem", ".key",
    "fixtures/private", "testdata",
    "config.dev", "dev.config", ".env"
)

$violations = @()
foreach ($entry in $entries) {
    foreach ($fragment in $forbidden) {
        if ($entry.ToLowerInvariant().Contains($fragment)) {
            $violations += "$entry (matches '$fragment')"
        }
    }
}

if ($violations.Count -gt 0) {
    $violations | ForEach-Object { Write-Error "forbidden in release: $_" -ErrorAction Continue }
    throw "inspect_release: $($violations.Count) forbidden entries found"
}

Write-Host "inspect_release: $($entries.Count) entries clean - no dev MCP, credentials, private fixtures or dev config."
