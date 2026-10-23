# Install the latest gitwhy release into %LOCALAPPDATA%\gitwhy\bin (or $env:GITWHY_BIN_DIR).
# Usage: irm https://raw.githubusercontent.com/dpologdvinity/gitwhy/main/install.ps1 | iex
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # Invoke-WebRequest is very slow with the progress bar
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$repo = 'dpologdvinity/gitwhy'
$binDir = if ($env:GITWHY_BIN_DIR) { $env:GITWHY_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'gitwhy\bin' }

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
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $binDir) {
  [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ";$binDir"), 'User')
  $env:Path += ";$binDir"
  Write-Host "Added $binDir to your PATH. Open a new terminal for other windows to see it."
}
