$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Source = Join-Path $ScriptDir "webreplica.exe"

if (-not (Test-Path $Source)) {
    throw "webreplica.exe was not found next to install.ps1"
}

$InstallDir = if ($env:WEBREPLICA_INSTALL_DIR) {
    $env:WEBREPLICA_INSTALL_DIR
} else {
    Join-Path $env:USERPROFILE "bin"
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$Target = Join-Path $InstallDir "webreplica.exe"
Copy-Item -Force $Source $Target

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$PathItems = @()
if ($UserPath) {
    $PathItems = $UserPath -split ";"
}

if ($PathItems -notcontains $InstallDir) {
    $NewPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
}

if (($env:Path -split ";") -notcontains $InstallDir) {
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "Installed webreplica to $Target"
Write-Host "You can now run:"
Write-Host "  webreplica https://example.com"
