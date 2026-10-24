'use strict';
// Downloads the gitwhy release binary for this OS/CPU (GitHub releases, checksum-verified).
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const https = require('node:https');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');

const REPO = 'dpologdvinity/gitwhy';
const OS = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const CPU = { x64: 'amd64', arm64: 'arm64' };

function version() {
  const v = process.env.GITWHY_VERSION || require('./package.json').version;
  if (v === '0.0.0') throw new Error('this package has no release version (set GITWHY_VERSION to test)');
  return v;
}

function target(ver, platform = process.platform, arch = process.arch) {
  const o = OS[platform];
  const c = CPU[arch];
  if (!o || !c) throw new Error(`gitwhy has no build for ${platform}/${arch}`);
  const archive = `gitwhy_${ver}_${o}_${c}.${o === 'windows' ? 'zip' : 'tar.gz'}`;
  const base = `https://github.com/${REPO}/releases/download/v${ver}`;
  return { archive, url: `${base}/${archive}`, checksums: `${base}/checksums.txt`, exe: o === 'windows' ? 'gitwhy.exe' : 'gitwhy' };
}

function expectedSum(checksums, archive) {
  for (const line of checksums.split('\n')) {
    const [sum, name] = line.trim().split(/\s+/);
    if (name === archive) return sum;
  }
  throw new Error(`no checksum for ${archive}`);
}

function verify(data, sum) {
  const got = crypto.createHash('sha256').update(data).digest('hex');
  if (got !== sum) throw new Error(`checksum mismatch (expected ${sum}, got ${got})`);
}

/**
 * Download a URL over HTTPS into memory, following up to `redirects` redirects.
 * Rejects on a non-200 status, a network error, or `timeoutMs` without any data
 * (a stalled connection would otherwise hang `npm install` forever).
 * @param {string} url
 * @param {number} [redirects=5]
 * @param {number} [timeoutMs=30000]
 * @returns {Promise<Buffer>}
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

// tarBin is the tar that can read our archive. On Windows, PATH may start with Git Bash's GNU tar,
// which can't read zip or C:\ paths; the system bsdtar (Windows 10+) can.
function tarBin(platform = process.platform, env = process.env) {
  return platform === 'win32' ? path.win32.join(env.SystemRoot || 'C:\\Windows', 'System32', 'tar.exe') : 'tar';
}

// place builds the install in a staging folder next to dir, then renames it into place, so a
// crash mid-extract never leaves a truncated binary that later runs would trust. If another
// process finished first, theirs is kept.
function place(dir, exe, produce) {
  fs.mkdirSync(path.dirname(dir), { recursive: true });
  const stage = fs.mkdtempSync(dir + '.tmp-');
  try {
    produce(stage);
    fs.chmodSync(path.join(stage, exe), 0o755);
    try {
      fs.renameSync(stage, dir);
    } catch (e) {
      if (!fs.existsSync(path.join(dir, exe))) throw e;
      fs.rmSync(stage, { recursive: true, force: true });
    }
  } catch (e) {
    fs.rmSync(stage, { recursive: true, force: true });
    throw e;
  }
}

// ensure returns the path of the gitwhy binary, downloading it on first use.
async function ensure() {
  const ver = version();
  const t = target(ver);
  const dir = path.join(__dirname, 'vendor', ver);
  const bin = path.join(dir, t.exe);
  if (fs.existsSync(bin)) return bin;

  const [data, sums] = await Promise.all([get(t.url), get(t.checksums)]);
  verify(data, expectedSum(sums.toString(), t.archive));

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
