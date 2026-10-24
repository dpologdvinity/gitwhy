#!/usr/bin/env node
'use strict';
/**
 * @file The `gitwhy` command that npm puts on PATH (package.json "bin").
 *
 * Makes sure the real binary is downloaded (lib.js ensure), then runs it with the same
 * arguments and terminal, and exits with its exit code. Reads GITWHY_VERSION through lib.js.
 */
const { spawnSync } = require('node:child_process');
const { ensure } = require('../lib');

ensure()
  .then((bin) => {
    // stdio 'inherit' gives gitwhy the real terminal, so its prompts and full-screen view work.
    const r = spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });
    if (r.error) throw r.error;
    // status is null when the binary was killed by a signal; report that as a failure.
    process.exit(r.status ?? 1);
  })
  .catch((e) => {
    console.error('gitwhy:', e.message);
    process.exit(1);
  });
