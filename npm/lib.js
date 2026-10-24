'use strict';
/**
 * @file Downloads the gitwhy release binary for this OS/CPU (GitHub releases, checksum-verified).
 *
 * Used by install.js (at `npm install`) and bin/gitwhy.js (on every run). The binary is kept in
 * vendor/<version>/ inside the package, so it is downloaded once per version. Needs Node 18+,
 * HTTPS access to github.com, and a `tar` that can read the archive (bsdtar on Windows 10+).
 *
 * Environment: GITWHY_VERSION overrides the release version from package.json (for testing).
 */
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const https = require('node:https');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');

const REPO = 'dpologdvinity/gitwhy';
// Node's process.platform and process.arch names, mapped to the GOOS and GOARCH names that
// goreleaser puts in the archive names.
const OS = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const CPU = { x64: 'amd64', arm64: 'arm64' };

/**
 * Return the gitwhy release version to download.
 * The release workflow sets package.json's version from the tag; the copy in git is 0.0.0, so
 * a package that was never released fails here with a clear message.
 * @returns {string} the version without a leading "v", e.g. "0.7.0"
 * @throws {Error} when the version is the 0.0.0 placeholder
 */
function version() {
  const v = process.env.GITWHY_VERSION || require('./package.json').version;
  if (v === '0.0.0') throw new Error('this package has no release version (set GITWHY_VERSION to test)');
  return v;
}

/**
 * Work out which release files to download for a platform. The names match goreleaser's
 * archives: a zip on Windows, a tar.gz elsewhere.
 * @param {string} ver release version without the "v"
 * @param {string} [platform=process.platform] Node platform name, e.g. "linux" or "win32"
 * @param {string} [arch=process.arch] Node CPU name, e.g. "x64" or "arm64"
 * @returns {{archive: string, url: string, checksums: string, exe: string}} the archive file
 *   name, its download URL, the checksums.txt URL, and the binary's name inside the archive
 * @throws {Error} when there is no release build for this platform and CPU
 */
function target(ver, platform = process.platform, arch = process.arch) {
  const o = OS[platform];
  const c = CPU[arch];
  if (!o || !c) throw new Error(`gitwhy has no build for ${platform}/${arch}`);
  const archive = `gitwhy_${ver}_${o}_${c}.${o === 'windows' ? 'zip' : 'tar.gz'}`;
  const base = `https://github.com/${REPO}/releases/download/v${ver}`;
  return { archive, url: `${base}/${archive}`, checksums: `${base}/checksums.txt`, exe: o === 'windows' ? 'gitwhy.exe' : 'gitwhy' };
}

/**
 * Find an archive's SHA-256 in the text of checksums.txt ("<sum>  <file name>" per line).
 * @param {string} checksums the whole checksums.txt
 * @param {string} archive the archive file name to look for
 * @returns {string} the hex SHA-256 listed for that file
 * @throws {Error} when the file is not listed
 */
function expectedSum(checksums, archive) {
  for (const line of checksums.split('\n')) {
    const [sum, name] = line.trim().split(/\s+/);
    if (name === archive) return sum;
  }
  throw new Error(`no checksum for ${archive}`);
}

/**
 * Check downloaded data against its expected SHA-256, so a changed or broken download is never
 * installed.
 * @param {Buffer} data the downloaded archive
 * @param {string} sum the expected hex SHA-256
 * @returns {void}
 * @throws {Error} when the hashes differ
 */
function verify(data, sum) {
  const got = crypto.createHash('sha256').update(data).digest('hex');
  if (got !== sum) throw new Error(`checksum mismatch (expected ${sum}, got ${got})`);
}

/**
 * Download a URL over HTTPS into memory, following up to `redirects` redirects.
 * Rejects on a non-200 status, a network error, or `timeoutMs` without any data
 * (a stalled connection would otherwise hang `npm install` forever).
 * @param {string} url the https URL to download
 * @param {number} [redirects=5] how many more redirects to follow (GitHub sends release
 *   downloads to a storage host)
 * @param {number} [timeoutMs=30000] how long to wait for any data before giving up
 * @returns {Promise<Buffer>} the whole response body
 */
function get(url, redirects = 5, timeoutMs = 30000) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, { headers: { 'user-agent': 'gitwhy-npm' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location && redirects > 0) {
        res.resume();
        return resolve(get(new URL(res.headers.location, url).href, redirects - 1, timeoutMs));
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`GET ${url}: HTTP ${res.statusCode}`));
      }
      const chunks = [];
      res.on('data', (d) => chunks.push(d));
      res.on('end', () => resolve(Buffer.concat(chunks)));
    });
    // An idle timeout: it fires only when no data arrives for timeoutMs, so slow but steady
    // downloads still finish. destroy() makes the request emit 'error' with our message.
    req.setTimeout(timeoutMs, () => req.destroy(new Error(`GET ${url}: no response for ${timeoutMs / 1000} s`)));
    req.on('error', reject);
  });
}

/**
 * Return the tar that can read our archive. On Windows, PATH may start with Git Bash's GNU tar,
 * which can't read zip or C:\ paths; the system bsdtar (Windows 10+) can.
 * @param {string} [platform=process.platform] Node platform name
 * @param {NodeJS.ProcessEnv} [env=process.env] environment, read for SystemRoot on Windows
 * @returns {string} the full path of System32\tar.exe on Windows, or "tar" (found on PATH)
 */
function tarBin(platform = process.platform, env = process.env) {
  return platform === 'win32' ? path.win32.join(env.SystemRoot || 'C:\\Windows', 'System32', 'tar.exe') : 'tar';
}

/**
 * Build the install in a staging folder next to dir, then rename it into place, so a crash
 * mid-extract never leaves a truncated binary that later runs would trust. If another process
 * finished first, theirs is kept. The staging folder is removed on every failure.
 * @param {string} dir final install folder, e.g. vendor/<version>
 * @param {string} exe binary file name inside the folder
 * @param {(stage: string) => void} produce writes the binary into the staging folder it is given
 * @returns {void}
 * @throws {Error} whatever produce throws, or a rename error when no finished install exists
 */
function place(dir, exe, produce) {
  fs.mkdirSync(path.dirname(dir), { recursive: true });
  const stage = fs.mkdtempSync(dir + '.tmp-');
  try {
    produce(stage);
    fs.chmodSync(path.join(stage, exe), 0o755);
    try {
      fs.renameSync(stage, dir);
    } catch (e) {
      // The rename fails when dir already exists. That is fine if it holds a finished binary.
      if (!fs.existsSync(path.join(dir, exe))) throw e;
      fs.rmSync(stage, { recursive: true, force: true });
    }
  } catch (e) {
    fs.rmSync(stage, { recursive: true, force: true });
    throw e;
  }
}

/**
 * Return the path of the gitwhy binary, downloading it on first use.
 * @returns {Promise<string>} absolute path of the binary in vendor/<version>/
 * @throws {Error} (as a rejection) on no release version, no build for this platform, a failed
 *   download, a checksum mismatch, or a failed extract
 */
async function ensure() {
  const ver = version();
  const t = target(ver);
  const dir = path.join(__dirname, 'vendor', ver);
  const bin = path.join(dir, t.exe);
  // place() only creates the binary once it is complete, so an existing one can be trusted.
  if (fs.existsSync(bin)) return bin;

  const [data, sums] = await Promise.all([get(t.url), get(t.checksums)]);
  verify(data, expectedSum(sums.toString(), t.archive));

  // tar needs a file, not a buffer, so write the verified archive to a temporary folder.
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'gitwhy-'));
  try {
    const file = path.join(tmp, t.archive);
    fs.writeFileSync(file, data);
    place(dir, t.exe, (stage) => execFileSync(tarBin(), ['-xf', file, '-C', stage, t.exe]));
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }
  return bin;
}

module.exports = { target, expectedSum, verify, get, tarBin, place, ensure };
