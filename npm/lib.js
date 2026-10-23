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

function get(url, redirects = 5) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'user-agent': 'gitwhy-npm' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location && redirects > 0) {
        res.resume();
        return resolve(get(new URL(res.headers.location, url).href, redirects - 1));
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`GET ${url}: HTTP ${res.statusCode}`));
      }
      const chunks = [];
      res.on('data', (d) => chunks.push(d));
      res.on('end', () => resolve(Buffer.concat(chunks)));
    }).on('error', reject);
  });
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
    fs.mkdirSync(dir, { recursive: true });
    // tar ships with Linux, macOS and Windows 10+ (bsdtar reads zip too).
    execFileSync('tar', ['-xf', file, '-C', dir, t.exe]);
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }
  fs.chmodSync(bin, 0o755);
  return bin;
}

module.exports = { target, expectedSum, verify, ensure };
