# git credential

Get, save, or forget credentials the way git does (config, helpers like keychains, or prompting). For scripts.

Reads a description from stdin: `key=value` lines, then a blank line.

**Usage:** `git credential (fill | approve | reject | capability) < <description>`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `fill` | Find username and password (helpers, then user). Prints the description with them added. |
| `approve` | Tell helpers it worked, so they may save it. |
| `reject` | Tell helpers it failed, so they may delete it. |
| `capability` | Print `version 0` and a `capability <name>` line per feature. |

Subcommands take no flags.

## Typical use

1. Write a description, e.g. `protocol=https`, `host=example.com`, `path=foo.git`, then a blank line.
2. Pipe it to `git credential fill`. Read back `username=` and `password=`.
3. Try the credential.
4. Pipe step 2's full output to `approve` if it worked, `reject` if not.

## Input/output keys

One `key=value` per line. Keys can't contain `=`, newline, or NUL; values can't contain newline or NUL. No quoting. Max 65535 bytes per line. Keys ending `[]` repeat as a list; empty `key[]=` clears it. Unknown keys ignored.

| Key | Meaning |
|-----|---------|
| `protocol` | e.g. `https`. |
| `host` | Hostname, plus port if any (`example.com:8088`). |
| `path` | Server path, e.g. repo. Dropped for HTTP(S) unless `credential.useHttpPath`. |
| `username` | Username, if known. |
| `password` | Password, when storing. |
| `password_expiry_utc` | Expiry, Unix seconds. `fill` skips expired. |
| `oauth_refresh_token` | OAuth refresh token stored with the password. Secret. |
| `url` | Split into `protocol`, `host`, etc. Protocol required. |
| `authtype` | Auth scheme, e.g. `basic`, `bearer` (`digest` insecure). Needs `capability[]=authtype`. |
| `credential` | Ready credential for the scheme, replacing username/password. HTTP sends `Authorization: <authtype> <credential>`. Needs `authtype` capability. |
| `ephemeral` | `true`: short-lived, don't save `credential`. Default false. Needs `authtype` capability. |
| `state[]` | Opaque helper data returned on its next call. Prefix uniquely. Needs `state` capability. |
| `continue` | Set by helper when a multistage login (NTLM, Kerberos) continues. Needs `state` capability. |
| `wwwauth[]` | Each server `WWW-Authenticate` header, in order. Git to helper only. |
| `capability[]` | Announce `authtype` or `state` support. Must come first. |
