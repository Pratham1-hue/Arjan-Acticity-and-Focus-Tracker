# Register com.studytracker as a Native Messaging host for Chrome and Edge (current user only).
# Run once after building, and re-run if you move the project folder.
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$manifest = Join-Path $PWD "host\com.studytracker.json"
if (-not (Test-Path $manifest)) { throw "manifest not found: $manifest" }

$keys = @(
  "HKCU:\Software\Google\Chrome\NativeMessagingHosts\com.studytracker",
  "HKCU:\Software\Microsoft\Edge\NativeMessagingHosts\com.studytracker"
)

foreach ($k in $keys) {
  New-Item -Path $k -Force | Out-Null
  Set-ItemProperty -Path $k -Name "(Default)" -Value $manifest
  Write-Host "registered: $k -> $manifest"
}

Write-Host ""
Write-Host "IMPORTANT: open host\com.studytracker.json and replace"
Write-Host "REPLACE_WITH_YOUR_EXTENSION_ID with the ID shown at"
Write-Host "chrome://extensions (after you load the extension in Phase 2)."
