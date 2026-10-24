<#
.SYNOPSIS
  Install the latest gitwhy release on Windows.

.DESCRIPTION
  Downloads the latest gitwhy release zip for this CPU (amd64 or arm64) from GitHub, checks its
  SHA-256 against the release's checksums.txt, and copies gitwhy.exe into
  %LOCALAPPDATA%\gitwhy\bin (or $env:GITWHY_BIN_DIR). Adds that folder to your user PATH if it
  is not there yet. Nothing is installed if the checksum does not match.

.EXAMPLE
  irm https://raw.githubusercontent.com/dpologdvinity/gitwhy/main/install.ps1 | iex

.EXAMPLE
  $env:GITWHY_BIN_DIR = 'C:\tools'; ./install.ps1

.NOTES
  Requires: Windows PowerShell 5.1 or PowerShell 7, and internet access to github.com.
  Environment: GITWHY_BIN_DIR sets the install folder.
  CI runs this script under both PowerShell versions (.github/workflows/ci.yml).
#>
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # Invoke-WebRequest is very slow with the progress bar
# Windows PowerShell 5.1 may default to TLS 1.0, which GitHub refuses.
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$repo = 'dpologdvinity/gitwhy'
$binDir = if ($env:GITWHY_BIN_DIR) { $env:GITWHY_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'gitwhy\bin' }

# A 32-bit PowerShell on 64-bit Windows reports x86 here; PROCESSOR_ARCHITEW6432 has the real CPU.
$cpu = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$arch = switch ($cpu) {
  'AMD64' { 'amd64' }
  'ARM64' { 'arm64' }
  default { throw "gitwhy: unsupported CPU $cpu" }
}

# The "latest" release URL redirects to .../tag/vX.Y.Z.
$req = [Net.HttpWebRequest]::Create("https://github.com/$repo/releases/latest")
$req.AllowAutoRedirect = $false
$resp = $req.GetResponse()
$tag = ([string]$resp.Headers['Location']) -replace '.*/tag/', ''
$resp.Close()
if ($tag -notmatch '^v\d') { throw "gitwhy: couldn't find the latest release" }
$version = $tag.Substring(1)
$file = "gitwhy_${version}_windows_$arch.zip"
$base = "https://github.com/$repo/releases/download/$tag"

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("gitwhy-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Invoke-WebRequest -UseBasicParsing "$base/$file" -OutFile (Join-Path $tmp $file)
  Invoke-WebRequest -UseBasicParsing "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

  # Verify before installing: the line for our zip in checksums.txt must match its SHA-256.
  $want = (Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match "\s$([regex]::Escape($file))$" } | Select-Object -First 1) -split '\s+' | Select-Object -First 1
  $got = (Get-FileHash -Algorithm SHA256 (Join-Path $tmp $file)).Hash.ToLower()
  if (-not $want -or $want -ne $got) { throw "gitwhy: checksum mismatch for $file; not installing" }

  Expand-Archive -Path (Join-Path $tmp $file) -DestinationPath $tmp -Force
  New-Item -ItemType Directory -Path $binDir -Force | Out-Null
  Copy-Item (Join-Path $tmp 'gitwhy.exe') (Join-Path $binDir 'gitwhy.exe') -Force
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

Write-Host "Installed gitwhy $version to $(Join-Path $binDir 'gitwhy.exe')"
# The user Path can be missing (null) or have empty entries (";;" or a trailing ";"). Split it,
# drop the empty parts, and join again, so we never write a leading, double or trailing ";".
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$parts = @(if ($userPath) { $userPath -split ';' | Where-Object { $_ } })
if ($parts -notcontains $binDir) {
  [Environment]::SetEnvironmentVariable('Path', (($parts + $binDir) -join ';'), 'User')
  $env:Path = (@($env:Path -split ';' | Where-Object { $_ }) + $binDir) -join ';'
  Write-Host "Added $binDir to your PATH. Open a new terminal for other windows to see it."
}
