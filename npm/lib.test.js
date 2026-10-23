'use strict';
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
