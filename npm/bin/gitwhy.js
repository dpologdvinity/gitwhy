#!/usr/bin/env node
'use strict';
const { spawnSync } = require('node:child_process');
const { ensure } = require('../lib');

ensure()
  .then((bin) => {
    const r = spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });
    if (r.error) throw r.error;
    process.exit(r.status ?? 1);
  })
  .catch((e) => {
    console.error('gitwhy:', e.message);
    process.exit(1);
  });
