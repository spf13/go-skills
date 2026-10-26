# Contributing

This repository publishes six [Agent Skills](https://docs.claude.com/en/docs/claude-code/skills) for Go as a Claude Code marketplace. A contribution is usually one of:

- a new skill,
- a fix or expansion to an existing skill's `SKILL.md`,
- a change to the catalog metadata.

Every structural rule below is enforced by CI. Contribution shape is checked on every push and pull request by `.github/workflows/validate.yml`; the Go release currency rule runs on a schedule in `.github/workflows/go-release-check.yml`. If a rule is not in this document, CI does not check it; if it is in this document, CI fails when you break it — with one exception, the description standard in §SKILL.md, which is judged in review because no check reads prose yet.

## Repository shape

One directory per skill at the repository root. Each directory is a single-skill Claude Code plugin:

```text
<skill>/
  SKILL.md
  .claude-plugin/
    plugin.json
```

- `SKILL.md` sits at the plugin root, not inside `.claude-plugin/`. Claude Code takes the invocation name from the frontmatter `name`.
- The catalog-level `.claude-plugin/marketplace.json` lists every skill. The six skill directories and the six marketplace entries must correspond one to one — no orphan directories, no entries pointing at missing directories.
- `templates/` is scaffold, not a skill; it is excluded from the correspondence check.
- The skills table in `README.md` lists every skill exactly once, each row linking to `./<skill>/SKILL.md`.

## Manifest contract

`<skill>/.claude-plugin/plugin.json` is the **single source of truth** for a skill's display metadata: `name`, `displayName`, `description`, `version`, `author`, `homepage`, `repository`, `license`, `keywords`.

The matching `marketplace.json` entry carries **exactly four keys**:

```json
{ "name": "<skill>", "source": "./<skill>", "category": "<taxonomy>", "tags": ["..."] }
```

Do not repeat `description`, `version`, `author`, `homepage`, `repository`, `license`, or `keywords` in the entry. Claude Code reads them from `plugin.json`, and duplicated copies drift out of sync.

`name` must equal the directory name and `plugin.json.name`.

## Category taxonomy

`category` must be exactly one of `languages`, `development`, `frameworks`. Choose by what the skill teaches:

| Category | Use when the skill teaches | Currently |
| --- | --- | --- |
| `languages` | the Go language itself | `go`, `cobra-viper` |
| `development` | an engineering workflow or practice | `go-spec-reviewer`, `go-release`, `fileflow-pathologize` |
| `frameworks` | a third-party framework or runtime | `wails` |

## Tags

An entry's `tags` must equal its `plugin.json` `keywords` as a set, with no duplicates on either side. They are the same list written once.

## SKILL.md

Frontmatter requires `name` and `description`.

- `name` must equal the skill directory name and `plugin.json.name`.
- `description` is what drives model-invoked triggering, and it is the only part of a skill a model reads before deciding whether to load the rest. CI checks that it exists and is at least 40 characters long; its shape, wording, and ceiling below are judged in review.

### Description standard

Two authoring standards are in circulation, and they disagree about what a description is for:

- **`writing-skills`** (Anthropic-style authoring): third person, trigger-only, opens `Use when…`, and MUST NOT summarize the workflow.
- **`skill-optimizer`** (Agent Skills spec): `definition + capability + trigger`, capped at 1024 characters / ~150 routing tokens, with no coercive activation language and no marketing framing.

**This catalog adopts the trigger-only opening, under the `skill-optimizer` budget.** The disagreement is the opening. A description is a routing signal matched against a prompt, so its first words have to be the condition that selects the skill. The leading definition `skill-optimizer` asks for spends that budget restating what the body's `# <name>` heading already says, and in practice becomes the workflow summary both standards ban. Capability and coverage live in the body, which the model reads after it has already routed. The ceiling and the bans apply unchanged on top.

A description MUST:

- Open with `Use when` or `Use for`. Nothing precedes it — no noun phrase, no self-description, no credential.
- Name the concrete contexts, file types, APIs, and phrases a user would actually type: `.go`, `go.mod`, `wails.json`, "cut a release", "is this spec ready?". A condition a model cannot match against a prompt is not a trigger.
- Stay under 1024 characters, and aim for ~150 routing tokens — roughly 600 characters. The ceiling is the hard cap; the target is what keeps the trigger legible.
- Stay in the third person and describe the user's task, not the skill: "Use when the user asks to…", never "I teach…" or "you should…".

A description MAY:

- Close with one clause that narrows its scope or defers to a sibling skill.
- Carry the `current through Go X.Y` claim §Go release currency requires — as a trailing clause, never as the opening or as the bulk of the text.

A description MUST NOT:

- Use coercive activation: "even if they don't ask", "recommend them", "Recommend these libraries even when they aren't named", "always use this skill", "CRITICAL:", "you must use this skill". An invocation instruction is not a trigger, and a model obeys one on prompts that have nothing to do with the skill.
- Summarize the workflow or enumerate coverage: "Covers idiomatic Go — package design, error handling, interfaces, concurrency, testing", "detect the version and follow only that version's section". A model that reads the summary acts on the summary, so the description becomes a second, unmaintained copy of the body — the failure both standards name.
- Use author-credential framing: "Expert skill for…", "Expert Go programming skill authored by spf13 (former Go team lead, author of Cobra…)", "authored by". Who wrote a skill never decides whether it applies; the condition does.
- Use marketing framing: "the definitive…", "the best…", "best-in-class". This is a routing string, not a landing page.

A conforming description:

```yaml
description: >
  Use when a Go program is slow, allocation-heavy, or suspected of leaking — pprof CPU, heap,
  block, mutex and goroutine profiles, benchmark comparison with -benchmem, escape analysis via
  -gcflags=-m, and cutting allocation counts. Not for correctness bugs; use the go skill.
```

`templates/skill/SKILL.md` carries that example as a whole file. The six shipped skills predate this rule and none of them conforms: fix one where you are already editing it, and do not treat the current six as models.

## Versioning and release

The catalog carries **one version across all six plugins**. They move together; there is no per-skill version.

To cut a release:

1. Set `version` in every `<skill>/.claude-plugin/plugin.json` to the same new value.
2. Add a matching entry to `CHANGELOG.md`.
3. Commit, then create an annotated tag `vX.Y.Z` on that commit.

CI asserts that all six versions are identical, that a commit tagged `vX.Y.Z` carries version `X.Y.Z`, that the version is never behind the latest existing tag, and that `CHANGELOG.md` has an entry for it. A version-bump pull request therefore passes before the tag exists; the tagged commit is what must match exactly.

## Go release currency

A skill that teaches the Go language states the release it covers inside its `description` frontmatter prose, in the literal form:

```text
current through Go X.Y
```

For example, `go/SKILL.md` carries `current through Go 1.27` in its description. §SKILL.md requires that claim to sit in a trailing clause, after the triggers — the `description` opens with `Use when`, so the claim never occupies the routing position. The check extracts exactly that phrase from the `description` value — the regular expression `current through Go (\d+\.\d+)`, matched against the parsed frontmatter (folded `>` scalars included) — and compares the major/minor line against the latest stable Go release. No extra frontmatter key is added; the phrase lives in the existing prose.

- A skill with no claim passes; the claim is optional.
- A skill that makes a claim must name the latest stable release. Being behind fails loudly, naming the skill, the claimed version, and the actual latest.
- The claim moves only in the same change that adds the release's features. Do not bump it ahead of a release that has not shipped.
- The claim must not disappear from a skill that carries one. The set of known claiming skills is recorded in the check itself (`CLAIMING_SKILLS` in `.github/scripts/check_go_release.py`), because the repository keeps no other record of it.

`.github/workflows/go-release-check.yml` enforces this on the first of each month and on manual dispatch. It reads the latest stable release from `https://go.dev/dl/?mode=json` (entries with `"stable": true`; pre-releases are ignored), runs the check, and on failure opens an issue titled `Go release drift`, or comments on the existing open one instead — so a long-standing drift does not file a new issue every month.

## Run the checks locally

```bash
python3 -m pip install --user jsonschema

# 1. Manifests validate against the published schemastore schemas
python3 .github/scripts/validate_manifests.py

# 2. Cross-file invariants (the rules in this document)
python3 .github/scripts/check_invariants.py

# 3. Go release currency (pass the value the workflow fetched)
python3 .github/scripts/check_go_release.py "$(curl -fsS 'https://go.dev/dl/?mode=json' | jq -r '[.[] | select(.stable == true)][0].version')"

# 4. Links in README.md and CONTRIBUTING.md resolve
lychee README.md CONTRIBUTING.md
```

## Adding a new skill

1. Copy the scaffold: `cp -r templates/skill <your-skill>`. The scaffold is a worked example named `go-profiling`; every fact in it belongs to that example, not to your skill.
2. Rewrite `SKILL.md` — frontmatter `name`/`description` first, then the guidance. The description must follow §SKILL.md; the example's description shows the shape to keep.
3. Fill in `.claude-plugin/plugin.json`: `name`, `displayName`, `description` (display copy for the marketplace, not a trigger), `keywords`, and `version` set to the catalog's current version — the scaffold ships `0.0.0` as a placeholder, and CI rejects any version behind the latest `v*` tag, so a copied-and-ignored placeholder fails loudly.
4. Copy `templates/skill/marketplace-entry.json` into `.claude-plugin/marketplace.json`, replacing the example's `name`, `source`, `category`, and `tags`. Keep the four keys.
5. Add a row to the skills table in `README.md`.
6. Run the checks above.
