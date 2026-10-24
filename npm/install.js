'use strict';
/**
 * @file The package's postinstall script (package.json "scripts.postinstall").
 *
 * Best effort: fetch the binary at install time. If it fails (offline, blocked scripts),
 * bin/gitwhy.js downloads it on first run instead, so the install never fails.
 * Reads GITWHY_VERSION through lib.js.
 */
require('./lib').ensure().catch((e) => console.warn('gitwhy: will download on first run (' + e.message + ')'));
