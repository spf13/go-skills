# Go Skills (`spf13/go-skills`) — Agent Instructions

Source repository for spf13's Go Agent Skills marketplace. The skills table in `README.md` lists every skill.

## Repository Layout

- Each top-level skill directory contains a `SKILL.md` with YAML frontmatter (`name`, `description`) and a `.claude-plugin/plugin.json`.
- `.claude-plugin/marketplace.json` is the Claude Code marketplace manifest. Keep it in sync whenever a skill is added, renamed, or removed.
- `CONTRIBUTING.md` lists every rule CI enforces. Read it before changing a skill or a manifest.

## Checks

Run these before committing. CI runs the same set, plus a link check of `README.md` and `CONTRIBUTING.md`; a monthly workflow checks each skill's `current through Go X.Y` claim against the latest Go release (`check_go_release.py`, see `CONTRIBUTING.md`).

```bash
python3 -m unittest discover -s .github/scripts
python3 .github/scripts/check_invariants.py
python3 .github/scripts/validate_manifests.py
python3 .github/scripts/check_skill_code.py        # needs the latest stable Go on PATH
(cd .github/claims && go test ./...)
```

## Using these skills from a local clone

Skills are plain directories, so a local clone can serve any agent that reads a skills directory:

- Link each skill into the agent's skills directory (for example `~/.agents/skills/<skill>`), as the README describes. Edits to a linked `SKILL.md` take effect immediately on that machine.
- If you sync skills across machines with a linking tool, re-run its sync after **adding, renaming, or removing** a skill so new links are created and stale ones removed. Editing an existing skill needs no re-sync.
- Other machines pick up changes with `git pull --ff-only` in their clone, followed by the sync step when skills were added, renamed, or removed.
