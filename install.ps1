# Installation script for amazing-cli (Windows PowerShell)
# This script downloads and installs the latest version of amazing-cli

$ErrorActionPreference = "Stop"

# GitHub repository
$repo = "huajianxiaowanzi/amazing-cli"
$binaryName = "amazing.exe"

Write-Host "🚀 Installing amazing-cli..." -ForegroundColor Green

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }
Write-Host "Detected OS: Windows"
Write-Host "Detected Architecture: $arch"

# Get latest release
try {
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $latestRelease.tag_name
    Write-Host "Latest version: $version"
} catch {
    Write-Host "Failed to get latest release information" -ForegroundColor Red
    exit 1
}

# Construct download URL
$archiveName = "amazing-cli_Windows_$arch.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$version/$archiveName"

Write-Host "Downloading from: $downloadUrl"

# Create temp directory
$tempDir = Join-Path $env:TEMP "amazing-cli-install"
if (Test-Path $tempDir) {
    Remove-Item $tempDir -Recurse -Force
}
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    # Download
    $archivePath = Join-Path $tempDir $archiveName
    Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath

    # Extract
    Write-Host "Extracting..."
    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

    # Install to user's local bin directory
    $installDir = Join-Path $env:USERPROFILE "bin"
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir | Out-Null
    }

    Write-Host "Installing to $installDir..."
    $sourcePath = Join-Path $tempDir $binaryName
    $destPath = Join-Path $installDir $binaryName
    Copy-Item $sourcePath $destPath -Force

    Write-Host "✅ Installation complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Run " -NoNewline
    Write-Host "amazing" -ForegroundColor Green -NoNewline
    Write-Host " to start using the CLI"
    Write-Host ""

    # Check if install directory is in PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        Write-Host "⚠️  Warning: $installDir is not in your PATH" -ForegroundColor Yellow
        Write-Host "Adding $installDir to your PATH..."
        $newPath = "$userPath;$installDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "✅ Added to PATH. Please restart your terminal for changes to take effect." -ForegroundColor Green
    }
} catch {
    Write-Host "Installation failed: $_" -ForegroundColor Red
    exit 1
} finally {
    # Cleanup
    if (Test-Path $tempDir) {
        Remove-Item $tempDir -Recurse -Force
    }
}
