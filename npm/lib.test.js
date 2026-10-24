'use strict';
/**
 * @file Tests for lib.js. Run with `node --test npm/lib.test.js` (CI does, on Linux and Windows).
 * No network: the download test uses a local server that never answers.
 */
const test = require('node:test');
const assert = require('node:assert');
const crypto = require('node:crypto');
const { target, expectedSum, verify } = require('./lib');

test('target picks the release archive for each platform', () => {
  assert.deepStrictEqual(
    [target('0.7.0', 'linux', 'x64').archive, target('0.7.0', 'darwin', 'arm64').archive, target('0.7.0', 'win32', 'x64').archive],
    ['gitwhy_0.7.0_linux_amd64.tar.gz', 'gitwhy_0.7.0_darwin_arm64.tar.gz', 'gitwhy_0.7.0_windows_amd64.zip'],
  );
  const t = target('0.7.0', 'win32', 'arm64');
  assert.strictEqual(t.exe, 'gitwhy.exe');
  assert.strictEqual(t.url, 'https://github.com/dpologdvinity/gitwhy/releases/download/v0.7.0/gitwhy_0.7.0_windows_arm64.zip');
  assert.strictEqual(target('0.7.0', 'linux', 'x64').exe, 'gitwhy');
});

test('target rejects platforms with no build', () => {
  assert.throws(() => target('0.7.0', 'freebsd', 'x64'), /no build for freebsd\/x64/);
  assert.throws(() => target('0.7.0', 'linux', 'ia32'), /no build for linux\/ia32/);
});

test('expectedSum finds the archive in checksums.txt', () => {
  const sums = 'aaa111  gitwhy_0.7.0_linux_amd64.tar.gz\nbbb222  gitwhy_0.7.0_windows_amd64.zip\n';
  assert.strictEqual(expectedSum(sums, 'gitwhy_0.7.0_windows_amd64.zip'), 'bbb222');
  assert.throws(() => expectedSum(sums, 'gitwhy_0.7.0_darwin_arm64.tar.gz'), /no checksum/);
});

test('verify rejects data that does not match the checksum', () => {
  const data = Buffer.from('hello');
  const sum = crypto.createHash('sha256').update(data).digest('hex');
  assert.doesNotThrow(() => verify(data, sum));
  assert.throws(() => verify(Buffer.from('hellp'), sum), /checksum mismatch/);
});

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { tarBin, place } = require('./lib');

test('tarBin uses the Windows system tar, not Git Bash GNU tar from PATH', () => {
  assert.strictEqual(tarBin('win32', { SystemRoot: 'C:\\Windows' }), 'C:\\Windows\\System32\\tar.exe');
  assert.strictEqual(tarBin('linux', {}), 'tar');
  assert.strictEqual(tarBin('darwin', {}), 'tar');
});

test('place leaves no binary behind when extraction fails', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gw-test-'));
  const dir = path.join(root, 'vendor', '1.0.0');
  assert.throws(() => place(dir, 'gitwhy', (stage) => {
    fs.writeFileSync(path.join(stage, 'gitwhy'), 'half'); // truncated write, then a crash
    throw new Error('tar died');
  }), /tar died/);
  assert.strictEqual(fs.existsSync(dir), false);
  assert.deepStrictEqual(fs.readdirSync(path.join(root, 'vendor')), []); // no staging dir left either
  fs.rmSync(root, { recursive: true });
});

test('place moves a finished extraction into place', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gw-test-'));
  const dir = path.join(root, 'vendor', '1.0.0');
  place(dir, 'gitwhy', (stage) => fs.writeFileSync(path.join(stage, 'gitwhy'), 'whole'));
  assert.strictEqual(fs.readFileSync(path.join(dir, 'gitwhy'), 'utf8'), 'whole');
  fs.rmSync(root, { recursive: true });
});

test('place keeps the winner when another process installed first', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gw-test-'));
  const dir = path.join(root, 'vendor', '1.0.0');
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(path.join(dir, 'gitwhy'), 'first');
  place(dir, 'gitwhy', (stage) => fs.writeFileSync(path.join(stage, 'gitwhy'), 'second'));
  assert.strictEqual(fs.readFileSync(path.join(dir, 'gitwhy'), 'utf8'), 'first');
  assert.deepStrictEqual(fs.readdirSync(path.join(root, 'vendor')), ['1.0.0']);
  fs.rmSync(root, { recursive: true });
});

const net = require('node:net');
const { get } = require('./lib');

test('get gives up with a clear error when the server never answers', async () => {
  // A TCP server that accepts the connection and then says nothing, like a stalled network.
  // The TLS handshake never finishes, so only get's idle timeout can end the request.
  const server = net.createServer(() => {});
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const url = `https://127.0.0.1:${server.address().port}/x`;
  try {
    await assert.rejects(get(url, 0, 200), /GET https:\/\/127\.0\.0\.1:\d+\/x: no response for 0\.2 s/);
  } finally {
    server.close();
  }
});
