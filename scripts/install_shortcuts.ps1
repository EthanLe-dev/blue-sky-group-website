$ErrorActionPreference = 'Stop'
$commandDirectory = Join-Path $env:LOCALAPPDATA 'BlueSky\bin'
$launcher = Join-Path $commandDirectory 'bluesky.cmd'
$projectCommand = Join-Path $PSScriptRoot 'bluesky.cmd'
if ((Test-Path -LiteralPath $launcher) -and
    ((Get-Content -LiteralPath $launcher -Raw) -notmatch 'Managed by Blue Sky Group')) {
    throw 'An unrelated bluesky command already exists at the install path.'
}
New-Item -ItemType Directory -Path $commandDirectory -Force | Out-Null
@"
@echo off
rem Managed by Blue Sky Group.
call "$projectCommand" %*
exit /b %errorlevel%
"@ | Set-Content -LiteralPath $launcher -Encoding ascii

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $commandDirectory) {
    [Environment]::SetEnvironmentVariable('Path', ($commandDirectory + ';' + $userPath), 'User')
}
if (($env:Path -split ';') -notcontains $commandDirectory) {
    $env:Path = $commandDirectory + ';' + $env:Path
}

# Refresh PATH for new PowerShell tabs even when their parent terminal has an old PATH.
$profilePaths = @($PROFILE.CurrentUserAllHosts)
$documents = [Environment]::GetFolderPath('MyDocuments')
if ($documents) { $profilePaths += Join-Path $documents 'WindowsPowerShell\profile.ps1' }
$block = @'
# Blue Sky Group command shortcuts
$blueSkyCommandBin = Join-Path $env:LOCALAPPDATA 'BlueSky\bin'
if (($env:Path -split ';') -notcontains $blueSkyCommandBin) {
    $env:Path = $blueSkyCommandBin + ';' + $env:Path
}
# End Blue Sky Group command shortcuts
'@
foreach ($profilePath in ($profilePaths | Select-Object -Unique)) {
    $encoding = [Text.UTF8Encoding]::new($true)
    if (Test-Path -LiteralPath $profilePath) {
        $reader = [IO.StreamReader]::new($profilePath, $true)
        try {
            $content = $reader.ReadToEnd()
            $encoding = $reader.CurrentEncoding
        } finally { $reader.Dispose() }
        if ($content.Contains('# Blue Sky Group command shortcuts')) { continue }
    }
    New-Item -ItemType Directory -Path (Split-Path -Parent $profilePath) -Force | Out-Null
    [IO.File]::AppendAllText($profilePath, [Environment]::NewLine + $block + [Environment]::NewLine, $encoding)
}
Write-Host 'Installed bluesky run/off. Open a new PowerShell tab. No app was started.'
