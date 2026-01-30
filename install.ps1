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
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "It appears this repository doesn't have any releases yet." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Alternative installation methods:" -ForegroundColor Cyan
    Write-Host "1. Install from source (requires Go):" -ForegroundColor White
    Write-Host "   go install github.com/$repo@latest" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2. Build from source:" -ForegroundColor White
    Write-Host "   git clone https://github.com/$repo.git" -ForegroundColor Gray
    Write-Host "   cd $($repo -replace '.*/')" -ForegroundColor Gray
    Write-Host "   go build -o $binaryName" -ForegroundColor Gray
    Write-Host ""
    Write-Host "For more information, visit: https://github.com/$repo" -ForegroundColor Cyan
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

    # Download and verify checksum
    Write-Host "Downloading checksums..."
    $checksumUrl = "https://github.com/$repo/releases/download/$version/checksums.txt"
    $checksumPath = Join-Path $tempDir "checksums.txt"
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath
        
        Write-Host "Verifying checksum..."
        $checksumContent = Get-Content $checksumPath | Where-Object { $_ -match $archiveName }
        if ($checksumContent) {
            $expectedHash = ($checksumContent -split '\s+')[0]
            $actualHash = (Get-FileHash $archivePath -Algorithm SHA256).Hash
            if ($expectedHash.ToLower() -ne $actualHash.ToLower()) {
                Write-Host "Checksum verification failed!" -ForegroundColor Red
                Write-Host "The downloaded file may be corrupted or tampered with." -ForegroundColor Red
                exit 1
            }
            Write-Host "Checksum verified successfully" -ForegroundColor Green
        }
    } catch {
        Write-Host "Warning: Could not verify checksum" -ForegroundColor Yellow
    }

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
        Write-Host ""
        Write-Host "The install directory is not in your PATH." -ForegroundColor Yellow
        $response = Read-Host "Would you like to add $installDir to your PATH? (y/N)"
        if ($response -eq 'y' -or $response -eq 'Y') {
            $newPath = "$userPath;$installDir"
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Host "✅ Added to PATH. Please restart your terminal for changes to take effect." -ForegroundColor Green
        } else {
            Write-Host "⚠️  You'll need to add $installDir to your PATH manually" -ForegroundColor Yellow
        }
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
