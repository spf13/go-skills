# CodeBuddy Skills

Project-local CodeBuddy skill discovery entry for the six skills in this marketplace.

## Layout

```
.codebuddy/skills/
├── README.md                        # this file
├── go/            -> ../../go
├── cobra-viper/   -> ../../cobra-viper
├── go-spec-reviewer/ -> ../../go-spec-reviewer
├── go-release/    -> ../../go-release
├── wails/         -> ../../wails
└── fileflow-pathologize/ -> ../../fileflow-pathologize
```

Each entry is a symlink to the canonical skill directory at the repo root — single source of truth, no content duplication. Edits happen in the root `*/SKILL.md`; this directory only exists so CodeBuddy's project-skill scanner picks the skills up at the path it expects.

## How CodeBuddy discovers a skill here

CodeBuddy's skill format requires only a `SKILL.md` per skill — no other file is mandatory. The per-skill `SKILL.md` in each symlinked directory already carries the required YAML frontmatter (`name`, `description`) and meets the activation-trigger standard from `https://www.codebuddy.cn/docs/ide/Features/Skills`.

If you ever need per-skill bundled resources (`scripts/`, `references/`, `assets/`), put them **next to the symlink target** at the repo root — that is the canonical skill directory. Do not place them inside `.codebuddy/skills/<name>/`; the symlink would shadow them.

## Editing a skill

Edit the canonical file at the repo root, not the symlink:

```
# edit the real file
$EDITOR go/SKILL.md

# symlink in .codebuddy/skills/ follows automatically
```

The marketplace catalog at `.claude-plugin/marketplace.json` and the per-skill `<skill>/.claude-plugin/plugin.json` must stay in sync with the symlinked `SKILL.md` (matching `name` and `description`).
