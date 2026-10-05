# Make Arjan start in the tray when you sign in to Windows.
# (build.ps1 already does this; the tray menu has a
# "Start with Windows" toggle too.)
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$exe = Join-Path $PWD "dist\tracker.exe"
if (-not (Test-Path $exe)) { throw "not built yet: $exe (run build.ps1 first)" }

Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "Arjan" -Value "`"$exe`""
Unregister-ScheduledTask -TaskName StudyTracker -Confirm:$false -ErrorAction SilentlyContinue

Write-Host "Arjan will now start in the tray at every sign-in."
Write-Host "To undo: tray icon -> uncheck 'Start with Windows'"
Write-Host "   (or: Remove-ItemProperty HKCU:\Software\Microsoft\Windows\CurrentVersion\Run -Name Arjan)"
