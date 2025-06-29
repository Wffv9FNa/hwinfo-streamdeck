# Check build prerequisites for HWiNFO Stream Deck Plugin
# This script verifies that all required dependencies are installed and properly configured

$ErrorActionPreference = 'Stop'
$requiredGoVersion = '1.22'

Write-Host "`n🔍 Checking build prerequisites for HWiNFO Stream Deck Plugin...`n" -ForegroundColor Cyan

function Test-Command {
    param (
        [string]$Name,
        [string]$Command,
        [string]$ExpectedPattern,
        [string]$ErrorMessage,
        [switch]$IsVersionCheck,
        [string]$MinVersion
    )

    Write-Host "Checking $Name... " -NoNewline

    try {
        $output = Invoke-Expression $Command
        if ($output -match $ExpectedPattern) {
            if ($IsVersionCheck -and $MinVersion) {
                $foundVersion = $matches[1]
                if ([version]$foundVersion -ge [version]$MinVersion) {
                    Write-Host "✅ Found v$foundVersion" -ForegroundColor Green
                    return $true
                } else {
                    Write-Host "❌ Version v$foundVersion is older than required v$MinVersion" -ForegroundColor Red
                    Write-Host $ErrorMessage -ForegroundColor Yellow
                    return $false
                }
            } else {
                Write-Host "✅ Found $($matches[0])" -ForegroundColor Green
                return $true
            }
        } else {
            Write-Host '❌ Version mismatch or not found' -ForegroundColor Red
            Write-Host $ErrorMessage -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host '❌ Not found' -ForegroundColor Red
        Write-Host $ErrorMessage -ForegroundColor Yellow
        return $false
    }
}

# Initialize status flags
$hasErrors = $false

# Check Go installation
$goCheck = Test-Command -Name 'Go' `
    -Command 'go version' `
    -ExpectedPattern 'go version go(\d+\.\d+\.\d+)' `
    -ErrorMessage "Please install Go $requiredGoVersion or later from https://go.dev/dl/" `
    -IsVersionCheck `
    -MinVersion $requiredGoVersion
if (-not $goCheck) { $hasErrors = $true }

# Check Git installation
$gitCheck = Test-Command -Name 'Git' `
    -Command 'git --version' `
    -ExpectedPattern 'git version (\d+\.\d+\.\d+)' `
    -ErrorMessage 'Please install Git from https://git-scm.com/downloads'
if (-not $gitCheck) { $hasErrors = $true }

# Function to find Visual Studio installation
function Find-VisualStudio {
    $vsInstallations = @(
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Enterprise",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Professional",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Community",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Enterprise",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Professional",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Community",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools"
    )

    # Try vswhere first
    $vswhereExe = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
    if (Test-Path $vswhereExe) {
        try {
            $installations = & $vswhereExe -prerelease -legacy -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
            if ($installations) {
                $vsInstallations = @($installations) + $vsInstallations
            }
        } catch {
            Write-Verbose "vswhere.exe failed: $_"
        }
    }

    # Filter out empty paths and check each valid path
    $vsInstallations = $vsInstallations | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    foreach ($vsPath in $vsInstallations) {
        if (-not [string]::IsNullOrWhiteSpace($vsPath) -and (Test-Path $vsPath)) {
            $vcvarsBat = Join-Path $vsPath 'VC\Auxiliary\Build\vcvars64.bat'
            if (Test-Path $vcvarsBat) {
                return $vcvarsBat
            }
        }
    }

    # Try Developer Command Prompt if available
    $devPrompt = @(
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Enterprise\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Professional\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Community\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Enterprise\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Professional\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\Community\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools\Common7\Tools\VsDevCmd.bat"
    ) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) -and (Test-Path $_) } | Select-Object -First 1

    if ($devPrompt) {
        return $devPrompt
    }

    return $null
}

# Check Visual Studio Build Tools
Write-Host 'Checking Visual Studio Build Tools... ' -NoNewline
$vcvarsBat = Find-VisualStudio
if ($vcvarsBat) {
    # Create a temporary batch file to capture environment variables
    $tempFile = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName() + '.bat')
    $envFile = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName() + '.txt')

    try {
        # Create batch file to call vcvars64.bat and save environment
        $setupScript = if ($vcvarsBat -match 'VsDevCmd\.bat$') {
            # If using VsDevCmd.bat, we need to specify amd64 architecture
            "@echo off`r`ncall `"$vcvarsBat`" -arch=amd64 -host_arch=amd64 > nul 2>&1`r`nif errorlevel 1 exit /b %errorlevel%`r`nset > `"$envFile`""
        } else {
            "@echo off`r`ncall `"$vcvarsBat`" > nul 2>&1`r`nif errorlevel 1 exit /b %errorlevel%`r`nset > `"$envFile`""
        }

        $setupScript | Out-File -FilePath $tempFile -Encoding ASCII

        # Execute the batch file
        $result = cmd /c "`"$tempFile`" 2>&1"
        $exitCode = $LASTEXITCODE

        # Clean up the batch file immediately
        Remove-Item $tempFile -ErrorAction SilentlyContinue

        if ($exitCode -ne 0) {
            throw "Visual Studio environment setup failed with exit code $exitCode"
        }

        if (-not (Test-Path $envFile)) {
            throw 'Environment variables file was not created'
        }

        # Read and parse the environment variables
        $envContent = Get-Content $envFile -ErrorAction Stop
        Remove-Item $envFile -ErrorAction SilentlyContinue

        $envContent | ForEach-Object {
            if ($_ -match '^([^=]+)=(.*)$') {
                $varName = $matches[1]
                $varValue = $matches[2]
                if ($varName -eq 'PATH') {
                    $env:PATH = $varValue
                }
            }
        }

        # Now try to run cl.exe with updated PATH
        try {
            $clOutput = cmd /c 'cl.exe 2>&1'
            if ($clOutput -match 'Microsoft \(R\) C/C\+\+ .*Compiler') {
                $version = $clOutput -split "`n" | Select-Object -First 1
                Write-Host "✅ Found $version" -ForegroundColor Green
            } else {
                throw "Compiler check failed: $clOutput"
            }
        } catch {
            Write-Host '❌ Found Visual Studio but C++ tools not installed or not configured' -ForegroundColor Red
            Write-Host @'
Please install Visual Studio 2022 C++ Build Tools:
1. Download from: https://visualstudio.microsoft.com/visual-cpp-build-tools/
2. In the installer, select "Desktop Development with C++"
3. Ensure these components are selected:
   - MSVC v143 - VS 2022 C++ x64/x86 build tools
   - Windows 10/11 SDK
   - C++ CMake tools for Windows
'@ -ForegroundColor Yellow
            $hasErrors = $true
        }
    } catch {
        Write-Host '❌ Failed to set up Visual Studio environment' -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
        $hasErrors = $true

        # Clean up any remaining temporary files
        Remove-Item $tempFile -ErrorAction SilentlyContinue
        Remove-Item $envFile -ErrorAction SilentlyContinue
    }
} else {
    Write-Host '❌ Visual Studio 2022 not found' -ForegroundColor Red
    Write-Host @'
Please install Visual Studio 2022 C++ Build Tools:
1. Download from: https://visualstudio.microsoft.com/visual-cpp-build-tools/
2. In the installer, select "Desktop Development with C++"
3. Ensure these components are selected:
   - MSVC v143 - VS 2022 C++ x64/x86 build tools
   - Windows 10/11 SDK
   - C++ CMake tools for Windows
'@ -ForegroundColor Yellow
    $hasErrors = $true
}

# Check Stream Deck software
$streamDeckPath = "$env:ProgramFiles\Elgato\StreamDeck\StreamDeck.exe"
Write-Host 'Checking Stream Deck Software... ' -NoNewline
if (Test-Path $streamDeckPath) {
    $version = (Get-Item $streamDeckPath).VersionInfo.FileVersion
    if ([version]$version -ge [version]'6.0') {
        Write-Host "✅ Found v$version" -ForegroundColor Green
    } else {
        Write-Host "❌ Version $version (requires ≥6.0)" -ForegroundColor Red
        Write-Host 'Please update Stream Deck software to version 6.0 or later' -ForegroundColor Yellow
        $hasErrors = $true
    }
} else {
    Write-Host '❌ Not found' -ForegroundColor Red
    Write-Host 'Please install Stream Deck software ≥v6.0 from https://www.elgato.com/downloads' -ForegroundColor Yellow
    $hasErrors = $true
}

# Check PowerShell version (optional)
$psCheck = Test-Command -Name 'PowerShell 7 (optional)' `
    -Command '$PSVersionTable.PSVersion.ToString()' `
    -ExpectedPattern '7\.(\d+\.\d+)' `
    -ErrorMessage 'PowerShell 7 is recommended but not required. Install from https://github.com/PowerShell/PowerShell/releases'

# Check protoc (optional)
$protocCheck = Test-Command -Name 'protoc (optional)' `
    -Command 'protoc --version' `
    -ExpectedPattern 'libprotoc (\d+\.\d+\.\d+)' `
    -ErrorMessage 'protoc is only needed when editing .proto files. The build system will download it automatically if needed.'

# Check HWiNFO installation
$hwInfoPath = "${env:ProgramFiles}\HWiNFO64\HWiNFO64.exe"
Write-Host 'Checking HWiNFO... ' -NoNewline
if (Test-Path $hwInfoPath) {
    Write-Host '✅ Found' -ForegroundColor Green

    # Check shared memory support
    Write-Host 'Checking HWiNFO Shared Memory Support... ' -NoNewline
    try {
        $handle = [System.IO.MemoryMappedFiles.MemoryMappedFile]::OpenExisting('Global\HWiNFO_SENS_SM2')
        $handle.Dispose()
        Write-Host '✅ Enabled' -ForegroundColor Green
    } catch {
        Write-Host '❌ Not enabled' -ForegroundColor Red
        Write-Host "Please enable 'Shared Memory Support' in HWiNFO Settings" -ForegroundColor Yellow
        $hasErrors = $true
    }
} else {
    Write-Host '❌ Not found' -ForegroundColor Red
    Write-Host 'Please install HWiNFO from https://www.hwinfo.com/download/' -ForegroundColor Yellow
    $hasErrors = $true
}

Write-Host "`nPrerequisite check completed!" -ForegroundColor Cyan
if ($hasErrors) {
    Write-Host '❌ Some required dependencies are missing. Please install them and run this script again.' -ForegroundColor Red
    exit 1
} else {
    Write-Host '✅ All required dependencies are installed and configured correctly.' -ForegroundColor Green
    exit 0
}
