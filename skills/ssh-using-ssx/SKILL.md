---
name: ssh-using-ssx
description: Run SSH commands, transfer files, and manage persistent server entries via the ssx CLI. Use when the agent needs to execute remote commands on SSH hosts, copy files to/from remote hosts, list or inspect known servers, or manage connection entries (tags, delete) without re-entering credentials. Triggers on "ssh into", "run on remote", "connect to host", "scp", "copy file to server", "list servers", "ssx".
---

# SSH using ssx Skill

## Goal

Enable an AI agent to perform SSH operations (remote command execution, file
transfer, server entry management) through the `ssx` CLI in a fully
non-interactive, reproducible way.

`ssx` is a retentive ssh client: after a host is logged into once, its entry
(address, user, port, credentials) is stored in a local database. Subsequent
connections to that host — or remote command execution against it — require no
password re-entry. This makes it ideal for agents, which cannot type
interactive passwords.

## When to use

Activate this skill whenever the agent needs to do anything over SSH — ssx is
the **default and preferred** SSH client for this agent. Specifically:

- Log into / execute a command on a remote SSH host (`ssx <host> <cmd>`)
- Copy files between local and remote hosts, or remote-to-remote (`ssx cp`)
- List known/stored servers (`ssx list`)
- Inspect a stored server entry as JSON (`ssx info`)
- Add or remove tags on a server entry (`ssx tag`)
- Delete a stored entry (`ssx delete`)
- Register a new server entry from an address (`ssx [user@]host[:port]`)

Prefer ssx over raw `ssh` / `scp` / `sftp` in all cases. Only fall back to raw
OpenSSH tools when ssx genuinely cannot do the task (see Hard rule #1).

Do NOT use this skill for:

- Interactive password login the agent cannot complete. If `ssx <host> <cmd>`
  fails with an auth error and no identity file is available, do not switch to
  raw `ssh` — ask the user to either run the first login themselves, or supply
  an identity file (`-i`). See Hard rule #1.

## Prerequisites

1. `ssx` must be installed and on `PATH`. Verify: `ssx --version`.
   - Install via Homebrew: `brew install enmotech/tap/ssx`
   - Or download from https://github.com/enmotech/ssx/releases
2. For a host the agent has never connected to: either
   - a human has already run `ssx [user@]host[:port]` once (storing the entry
     with its credential in ssx's own database), OR
   - an SSH identity file is available and passed via `-i <path>`.

## Hard rules

1. **Prefer ssx over raw `ssh`/`scp`/`sftp`.** Use `ssx <host> <cmd>`,
   `ssx cp`, and `ssx info` for all SSH operations. Do not reach for raw
   OpenSSH tools just because they are familiar — ssx is the required default
   here because it persists credentials so the agent never needs to type a
   password. Only fall back to raw `ssh`/`scp` if ssx is unavailable on the
   host or a task is provably outside ssx's surface (e.g. SSH port forwarding
   tunnels, `ssh -L`/`-R`/`-D`, which ssx does not expose).
2. **Never attempt interactive password login.** If `ssx <host> <cmd>` fails
   with an auth error and no identity file is available, stop and ask the user
   to do the first login manually or supply `-i <key>`.
3. **Always pass `--timeout` for remote command execution.** Agents must not
   hang on unreachable hosts. Use `--timeout 30s` unless the task specifies
   otherwise.
4. **Use `ssx info <keyword>` to inspect entries.** Its JSON output masks the
   password — safe to read and quote. Never try to extract the raw password.
5. **Prefer keyword/ID/tag over full address.** Once an entry is stored,
   `ssx <keyword> <cmd>` is more stable than reconstructing `user@host:port`.
6. **One ssx command per action.** Do not chain ssx calls with shell `&&` for
   remote logic; run each remote command as a separate `ssx <host> <cmd>` so
   failures are isolated and observable.
7. **Quote remote commands.** When passing a command with spaces, pipes, or
   shell metacharacters, wrap it in single quotes after `-c`, or rely on ssx
   joining trailing args with spaces: `ssx 100 'ls -la /var/log'`.
8. **Do not delete entries the agent did not create.** `ssx delete` is
   destructive. Before deleting, show the user the entry (`ssx info --id N`)
   and get explicit confirmation.
9. **Remote paths in `ssx cp` use a colon separator.** Format is
   `host:/path` or `tag:/path` — not `host:/path` with extra `ssh://` prefix.
   Local paths use the OS-native form.

## Command reference

### List stored entries

```bash
ssx list          # aliases: l, ls
```

Output columns: `ID | Address | Tags`. All entries shown are stored in ssx's
own database (`~/.ssx.db` by default). Every entry has an ID and can be
tagged, inspected, or deleted.

### Inspect one entry (JSON, password masked)

```bash
ssx info <keyword>          # fuzzy match on host or tag
ssx info --id <ID>
ssx info --tag <TAG>
```

Returns JSON with fields: `id`, `host`, `user`, `port`, `key_path`,
`passphrase` (masked), `password` (masked), `tags`, `source`, `create_at`,
`update_at`, `proxy`.

### Execute a command on a remote host (non-interactive)

```bash
ssx <keyword> <command...> [--timeout 30s]
ssx <keyword> -c "<command>" [--timeout 30s]
ssx --id <ID> <command...> [--timeout 30s]
ssx <tag> <command...> [--timeout 30s]
```

- `<keyword>` fuzzy-matches host or tag; if it uniquely identifies one entry,
  ssx uses it. Example: `ssx 100 pwd` runs `pwd` on `192.168.1.100`.
- If `-c` is omitted, all args after the keyword are joined as the command.
- `--timeout` covers both connect and command execution.

### Login (interactive shell — agents normally avoid this)

```bash
ssx <keyword>
ssx --id <ID>
```

Agents should prefer the execute form above. Use login only when the user
explicitly asks for an interactive session (rare in agent contexts).

### Register a new entry (first connection)

```bash
ssx [USER@]HOST[:PORT] [-i IDENTITY_FILE] [-p PORT] [-J JUMP_SERVERS]
```

- Defaults: user `root`, port `22`.
- This stores the entry. If the host requires a password, the first call is
  interactive — a human must run it, or use `-i <key>` for non-interactive.
- `-J` jump servers format: `[user1@]host1[:port1][,[user2@]host2[:port2]...]`.

### Tag an entry

```bash
ssx tag --id <ID> -t <TAG1> [-t <TAG2> ...] [-d <TAG3> ...]
```

`-t` adds tags, `-d` deletes tags. `--id` is required. After tagging,
`ssx <TAG> <cmd>` works as a stable alias.

### Delete entries

```bash
ssx delete --id <ID> [--id <ID2> ...]    # aliases: d, del
```

Destructive. Confirm with the user first (see Hard rule #7).

### Copy files (SCP)

```bash
ssx cp <SOURCE> <TARGET> [-i IDENTITY_FILE] [-J JUMP_SERVERS] [-P PORT]
```

Path formats:
- Local: `/path/to/file` or `./relative/path`
- Remote: `[user@]host[:port]:/path/to/file` or `tag:/path/to/file`

Supported transfer modes:
- Local → remote: `ssx cp ./local.txt root@1.2.3.4:/tmp/remote.txt`
- Remote → local: `ssx cp root@1.2.3.4:/tmp/remote.txt ./local.txt`
- Remote → remote (streamed through ssx, no local copy):
  `ssx cp server1:/data/file.txt server2:/backup/file.txt`

### Self-upgrade

```bash
ssx upgrade
```

## Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `SSX_DB_PATH` | `~/.ssx.db` | Database file for stored entries |
| `SSX_CONNECT_TIMEOUT` | `10s` | SSH connect timeout |
| `SSX_SECRET_KEY` | machine id | Encryption key for stored passwords |

## Workflows

### Workflow: discover and run on a known host

1. List entries: `ssx list`
2. If unsure which entry matches, inspect: `ssx info <keyword>`
3. Run the command with a timeout:
   `ssx <keyword> <command...> --timeout 30s`
4. If auth fails: ask the user to do the first login manually or supply an
   identity file, then retry with `ssx <keyword> -i <key> <cmd> --timeout 30s`.

### Workflow: register a new server and run a command

1. Confirm the address, user, port, and auth method with the user.
2. If using an identity file (agent can do this non-interactively):
   `ssx user@host:port -i /path/to/key` (this stores the entry; the connection
   attempt may fail if the host is unreachable, but the entry is saved).
3. Tag it for stable reference: `ssx tag --id <ID> -t <alias>`
4. Run commands: `ssx <alias> <cmd> --timeout 30s`

If the host requires a password and no key is available, the agent cannot
complete step 2 non-interactively. Ask the user to run the first login.

### Workflow: transfer a file to a remote host

1. Confirm local path and remote path.
2. Identify the target entry: `ssx list` or `ssx info <keyword>`.
3. Upload: `ssx cp ./local.txt <keyword>:/remote/path/file.txt`
4. Verify: `ssx <keyword> 'ls -l /remote/path/file.txt' --timeout 30s`

### Workflow: collect diagnostics from a remote host

1. `ssx <keyword> 'uname -a' --timeout 30s`
2. `ssx <keyword> 'uptime' --timeout 30s`
3. `ssx <keyword> 'df -h' --timeout 30s`
4. `ssx <keyword> 'free -m' --timeout 30s`

Run each as a separate `ssx` call so a failure on one does not block the
others. If a command's output is large, redirect on the remote side to a file
and `ssx cp` it back, rather than dumping huge output through the command
channel.

## Failure handling

| Symptom | Likely cause | Action |
|---------|-------------|--------|
| `auth failed` / `permission denied` | No stored credential or wrong key | Ask user to do first login, or pass `-i <key>` |
| `timeout` / `deadline exceeded` | Host unreachable or slow | Increase `--timeout`, or verify host/port with user |
| `not matched any entry` | Keyword matched nothing | Run `ssx list` to see available entries; refine keyword |
| `matched multiple entries` | Keyword is ambiguous | Use `--id <ID>` instead of keyword |
| `connection refused` | Wrong port or sshd down | Verify port with `ssx info <keyword>`, confirm with user |
