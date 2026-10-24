# Go Skills (`spf13/go-skills`) — Agent Instructions

Canonical repository for Steve Francia's (`spf13`) Go Agent Skills marketplace (`go`, `cobra-viper`, `go-spec-reviewer`, `go-release`, `wails`, `fileflow-pathologize`).

## Repository Layout

- Each top-level skill directory (`go/`, `cobra-viper/`, `go-spec-reviewer/`, `go-release/`, `wails/`, `fileflow-pathologize/`) contains a `SKILL.md` with YAML frontmatter (`name`, `description`) and optional `references/` or `scripts/`.
- `.claude-plugin/marketplace.json` defines the Claude Code plugin marketplace manifest. Keep it in sync whenever skills are added, renamed, or removed.

---

## Agent & Skill Ecosystem Integration (`dotagents` + `sage`)

### Architecture Map

- **`~/code/steve-skills`, `~/code/go-skills`, `~/code/cli-style`**: Canonical source repositories for Steve's authored skills (`D:\MyFiles\code\<repo>` on `galactus`, `/share/steve/code/<repo>` on `factory`, `~/code/<repo>` on `wandamax`).
- **`~/.agents` (`github.com/spf13/dotagents`, and `/share/steve/code/dotagents` in Orca on `factory`)**: Canonical multi-harness configuration repository (`sage.yaml`, `AGENTS.md`, `agents/`, `omp/`, `claude/`, `mcp/`).
- **`~/code/sage` (`github.com/spf13/sage`)**: CLI sync engine (`sage sync`) that links `~/code/<repo>/<skill>` into `~/.agents/skills/`, `~/.gemini/config/skills/`, and `~/.omp/agent/skills/`, and propagates `AGENTS.md` (`RULES.md`, `GEMINI.md`, `CLAUDE.md`), `agents/`, `omp/`, `claude/`, and `mcp/` across all harnesses.

### Mandatory Sync & Reload Protocol on Every Update

1. **Editing an existing skill (`SKILL.md`, `references/`, `scripts/`)**:
   - Skills are symlinked (`~/.agents/skills/<name> -> ~/code/go-skills/<name>`), so edits are live immediately on the current machine.
   - Commit & push to `origin/main`, then pull on remote hosts:
     ```bash
     ssh factory "git -C /share/steve/code/go-skills pull --ff-only"
     ssh wandamax "git -C ~/code/go-skills pull --ff-only"
     ```
2. **Adding, renaming, or removing a skill**:
   - Update `.claude-plugin/marketplace.json` if needed, then run `sage sync` locally (auto-discovers new skills in `~/code/go-skills` and records portable `local_path: ~/code/go-skills` in `~/.agents/sage.yaml`).
   - Commit & push both `go-skills` (`origin/main`) AND `~/.agents` (`spf13/dotagents` `origin/master`).
   - Sync remote hosts so new symlinks and manifests materialize everywhere:
     ```bash
     ssh factory "git -C /share/steve/code/go-skills pull --ff-only && git -C /root/.agents pull --ff-only && git -C /share/steve/code/dotagents pull --ff-only && sage sync"
     ssh wandamax "git -C ~/code/go-skills pull --ff-only && git -C ~/.agents pull --ff-only && ~/go/bin/sage sync"
     ```
3. **Updating `sage` itself (`~/code/sage`)**:
   - Run `go test ./... && go install .`, commit & push `origin/master`, then install on `factory` and `wandamax`:
     ```bash
     ssh factory "cd /share/steve/code/sage && git pull --ff-only && go install . && cp /root/go/bin/sage /usr/local/bin/sage"
     ssh wandamax "cd ~/code/sage && git pull --ff-only && go install ."
     ```
