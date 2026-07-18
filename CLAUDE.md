# nxtools — working agreement

This file travels with the repo, so the same rules apply on every machine I work
from (laptop + the two desktops), regardless of local Claude config.

## Permissions & behavior (author's standing instructions)

1. **Full read access, everywhere.** Reading is always granted — local filesystem
   (anywhere on disk, not just this repo), the web, etc. Never ask before reading.
   Enforced in `.claude/settings.json`: the `Read`/`Glob`/`Grep`/`WebFetch`/`WebSearch`
   tools are allow-listed and `additionalDirectories` is set to the filesystem root.

2. **Never run `git commit`.** Committing is the author's prerogative. I may draft
   and *propose* a commit message when asked, but I must not create commits myself.
   Enforced as a `deny` rule on `Bash(git commit:*)` in `.claude/settings.json`.
   (Branching/staging/other git ops are fine; only committing is off-limits.)

3. **Write access is confirmed once per session, then not re-asked.** Ask for write
   permission the first time it's needed; after it's granted, don't reconfirm for the
   rest of the session. There is no settings key that expresses "ask once then stop
   asking," so this is honored via the interactive prompt: choose **"Yes, and don't
   ask again this session"** on the first write. This preference is intentionally
   *not* encoded as a blanket allow, because that would mean "never ask" rather than
   "ask once."

## Notes
- `.claude/settings.json` is committed so it reaches every environment. Do not move
  these rules into `.claude/settings.local.json` — that file is gitignored and would
  not sync to the other machines.
