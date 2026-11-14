# Contributing

This repository publishes [Agent Skills](https://docs.claude.com/en/docs/claude-code/skills) for Go as a Claude Code marketplace. A contribution is usually one of:

- a new skill,
- a fix or expansion to an existing skill's `SKILL.md`,
- a change to the catalog metadata.

Every structural rule below is enforced by CI. Contribution shape is checked on every push and pull request by `.github/workflows/validate.yml`, which runs `go test ./...` and then `go run ./cmd/check`; the Go release currency rule runs on a schedule in `.github/workflows/go-release-check.yml`. The checker is one zero-dependency Go program (`cmd/check` + `internal/check`) that replaced three Python scripts: it needs no `pip install` and fetches nothing from the network. If a rule is not in this document, CI does not check it; if it is in this document, CI fails when you break it — with one exception, the wording half of the description standard in §SKILL.md, which is judged in review because no check reads prose meaning.

## Repository shape

One directory per skill at the repository root. Each directory is a single-skill Claude Code plugin:

```text
<skill>/
  SKILL.md
  .claude-plugin/
    plugin.json
```

- `SKILL.md` sits at the plugin root, not inside `.claude-plugin/`. Claude Code takes the invocation name from the frontmatter `name`.
- The catalog-level `.claude-plugin/marketplace.json` lists every skill. The skill directories and the marketplace entries must correspond one to one — no orphan directories, no entries pointing at missing directories.
- `templates/` is scaffold, not a skill; it is excluded from the correspondence check.
- The skills table in `README.md` lists every skill exactly once, each row linking to `./<skill>/SKILL.md`.

## Manifest contract

`<skill>/.claude-plugin/plugin.json` is the **single source of truth** for a skill's display metadata: `name`, `displayName`, `description`, `version`, `author`, `homepage`, `repository`, `license`, `keywords`.

Both manifests are validated against the published schemastore schemas (`claude-code-marketplace.json`, `claude-code-plugin-manifest.json`). Those schemas are vendored in `internal/check/schemas/` so the check stays hermetic; refresh them deliberately, never automatically:

```bash
curl -fsSL https://json.schemastore.org/claude-code-marketplace.json \
  -o internal/check/schemas/claude-code-marketplace.json
curl -fsSL https://json.schemastore.org/claude-code-plugin-manifest.json \
  -o internal/check/schemas/claude-code-plugin-manifest.json
```

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
| `development` | an engineering workflow or practice | `go-spec-reviewer`, `go-release`, `fileflow-pathologize`, `afero` |
| `frameworks` | a third-party framework or runtime | `wails` |

## Tags

An entry's `tags` must equal its `plugin.json` `keywords` as a set, with no duplicates on either side. They are the same list written once.

## SKILL.md

Frontmatter requires `name` and `description`.

- `name` must equal the skill directory name and `plugin.json.name`.
- `description` is what drives model-invoked triggering, and it is the only part of a skill a model reads before deciding whether to load the rest. CI checks that it exists, is at least 40 characters long, opens with `Use when` or `Use for`, stays under 1024 characters, and carries none of the banned coercive phrases below; wording beyond that is judged in review.
- The file stays within **500 lines**. The whole skill is one context load, so anything longer belongs in `references/`, which the model reads only when a task needs it.
- The body carries a `## When NOT to Use This Skill` section.
- No unresolved fill-in tokens such as `[PLACEHOLDER]` or `[SPEC_FILE_PATH]` outside fenced code: a model reads them as content.
- Every `references/…md` link resolves to a file.
- A fenced `go` sample that is complete — it carries its own `package` clause — must compile and vet. The check imports nothing, so only standard-library samples are built; a complete sample that imports another module is counted and skipped rather than failed. Partial samples (a statement, a signature, an excerpt) are never compiled.

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

`templates/skill/SKILL.md` carries that example as a whole file. The skills shipped before this rule did not all conform: fix one where you are already editing it, and do not treat the existing set as models.

Every skill body MUST carry a `## When NOT to Use This Skill` section naming the siblings it defers to, so that overlapping triggers resolve without a priority table. A new skill MUST include one, and an edit to an existing skill MUST add one where it is missing.

## Go code in skills

`.github/scripts/check_skill_code.py` runs every skill against the latest stable Go toolchain, using that toolchain's own `stdversion` manifest as the record of which release introduced each standard-library symbol:

- Every ```` ```go ```` fence, indented ones included, must parse with the toolchain's `gofmt`. Fragments are fine — a fence passes as a file, as declarations, as statements, or as declarations followed by statements, and `{ ... }` or a line of `...` marks elided code. Fence `go.mod` content and other non-Go text as ```` ```text ````.
- Every standard-library symbol a skill names must exist: backticked prose such as `errors.AsType`, `sync.WaitGroup.Go` or `t.Context()`, and `pkg.Name` selectors inside fences. Names a skill takes from a non-standard module — an import in a fence, or a module path it mentions such as `github.com/wailsapp/wails/v2/pkg/runtime` — are skipped unless the skill also imports the standard package of that name, as are names a fence declares itself as variables (in that same fence only). A lowercase name after a package (`errors.go`) is a file, not an API. A bare `Type.Method` is checked only when its standard package is named on the same line.
- Every standard-library version claim must name the right release. A claim is a `| 1.X | … |` row in a Since/Feature table, a heading ending in `(Go 1.X)`, or a backticked symbol written directly before `(Go 1.X)`. The claim's subject — its first symbol, or a package written as "the X package" — must have arrived exactly in 1.X; every other symbol must merely exist by then. A claim that records a behavior change ("`B.Loop` no longer inhibits inlining") has no subject, and a heading that names only a package as context ("`net/url` Clones") pins nothing. Language features have no symbol and are checked in review.
- A claim about runtime or library *behavior* — what a decoder accepts, what a channel's capacity is — gets a test in `.github/claims/` (`go test ./...` there), naming the `SKILL.md` section it backs. When Go changes the behavior, the test turns red and the skill text changes with it.

## Versioning and release

The catalog carries **one version across all plugins**. They move together; there is no per-skill version.

To cut a release:

1. Set `version` in every `<skill>/.claude-plugin/plugin.json` to the same new value.
2. Add a matching entry to `CHANGELOG.md`.
3. Commit, then create an annotated tag `vX.Y.Z` on that commit.

CI asserts that all plugin versions are identical, that a commit tagged `vX.Y.Z` carries version `X.Y.Z`, that the version is never behind the latest existing tag, and that `CHANGELOG.md` has an entry for it. A version-bump pull request therefore passes before the tag exists; the tagged commit is what must match exactly.

## Go release currency

A skill that teaches the Go language may state the release it covers inside its `description` frontmatter prose, in the literal form:

```text
current through Go X.Y
```

For example, `go/SKILL.md` carries `current through Go 1.27` in its description. §SKILL.md requires that claim to sit in a trailing clause, after the triggers — the `description` opens with `Use when`, so the claim never occupies the routing position. The check extracts exactly that phrase from the `description` value — the regular expression `current through Go (\d+\.\d+)`, matched against the parsed frontmatter (folded `>` scalars included) — and compares the major/minor line against the latest stable Go release. No extra frontmatter key is added; the phrase lives in the existing prose.

- A skill with no claim passes; the claim is optional.
- A skill that makes a claim must name the latest stable release. Being behind fails loudly, naming the skill, the claimed version, and the actual latest.
- The claim moves only in the same change that adds the release's features. Do not bump it ahead of a release that has not shipped.
- The claim must not disappear from a skill that carries one. The set of known claiming skills is recorded in the check itself (`claimingSkills` in `internal/check/release.go`), because the repository keeps no other record of it.

**Only `go` claims, by decision.** The phrase asserts coverage of the whole language and standard-library surface, and only `go` enumerates that surface. `go-release`, `go-spec-reviewer`, `cobra-viper`, and `fileflow-pathologize` teach individual 1.26/1.27 features and state the floor where each one applies (`Go 1.26+`, `go 1.27 or later`), so each goes stale feature-by-feature rather than release-by-release — a coverage claim on them would be false the moment one skill missed a release's feature. They stay claim-free. A skill that later asserts release coverage adds itself to `claimingSkills` in the same change that adds the claim; never add the claim to a skill that only consumes individual features.

`.github/workflows/go-release-check.yml` enforces this on the first of each month and on manual dispatch. It is the only part of the pipeline that touches the network: it reads the latest stable release from `https://go.dev/dl/?mode=json` (entries with `"stable": true`; pre-releases are ignored) and hands it to the checker as `-latest`. On failure it opens an issue titled `Go release drift`, or comments on the existing open one instead — so a long-standing drift does not file a new issue every month.

## Run the checks locally

```bash
go test ./...               # the checker's own tests
go run ./cmd/check          # every rule, against the working tree (hermetic, no network)

# Go release currency, with the value the scheduled workflow fetched:
go run ./cmd/check -latest go1.27.3    # or GO_RELEASE_LATEST=go1.27.3 go run ./cmd/check

# The skill-code checker's own tests, then every Go fence and std version claim
# (both need the latest stable Go on PATH):
python3 -m unittest discover -s .github/scripts
python3 .github/scripts/check_skill_code.py
(cd .github/claims && go test ./...)

# Links in README.md and CONTRIBUTING.md resolve
lychee README.md CONTRIBUTING.md
```

`go run ./cmd/check` exits 0 when every rule holds, 1 with a `path:line:` list of violations, and 2 when the checker itself cannot run. Run it from anywhere inside the repository; `-root` points it elsewhere. Without `-latest` the release comparison is skipped, because a local run has no network: the claim-must-not-disappear rule still applies.

## Adding a new skill

1. Copy the scaffold: `cp -r templates/skill <your-skill>`. The scaffold is a worked example named `go-profiling`; every fact in it belongs to that example, not to your skill.
2. Rewrite `SKILL.md` — frontmatter `name`/`description` first, then the guidance. The description must follow §SKILL.md; the example's description shows the shape to keep.
3. Fill in `.claude-plugin/plugin.json`: `name`, `displayName`, `description` (display copy for the marketplace, not a trigger), `keywords`, and `version` set to the catalog's current version — the scaffold ships `0.0.0` as a placeholder, and CI rejects any version behind the latest `v*` tag, so a copied-and-ignored placeholder fails loudly.
4. Copy `templates/skill/marketplace-entry.json` into `.claude-plugin/marketplace.json`, replacing the example's `name`, `source`, `category`, and `tags`. Keep the four keys.
5. Add a row to the skills table in `README.md`.
6. Run the checks above.
