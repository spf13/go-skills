# Baseline — 2026-11-04 @ 9adc400 (branch claude/festive-curie-d6zqps)

Repo has no compiled code. The product is six SKILL.md files plus manifests; the
Python scripts under `.github/scripts/` are CI guardrails.

| Check | Command | Result | Time |
| --- | --- | --- | --- |
| Manifest schema | `python3 .github/scripts/validate_manifests.py` | PASS, 7 manifests | 3.0s (network) |
| Invariants | `python3 .github/scripts/check_invariants.py` | PASS, 6 skills, v1.1.0 | 0.08s |
| Go release currency | `python3 .github/scripts/check_go_release.py go1.27.3` | PASS, 1 claim | <0.1s |
| Go release currency (red demo) | `... check_go_release.py go1.28.0` | FAIL as designed | <0.1s |
| Claude Code validate | `claude plugin validate .` and `./<skill>` (CLI 2.1.318) | PASS | ~1s |
| Links | `lychee README.md CONTRIBUTING.md` | NOT RUN: lychee not installed | — |
| Unit tests | — | NONE EXIST (0 tests for 3 scripts, 389 lines) | — |
| Lint/format | — | NONE CONFIGURED | — |
| Coverage | — | n/a | — |

## Content metrics
- SKILL.md sizes (bytes): go 50387 (1188 lines), cobra-viper 21355, go-release 17210,
  fileflow-pathologize 14196, go-spec-reviewer 12958, wails 9380. Total 125486.
- Go code fences: 88. Naive parse check: 69/88 parse; the 19 misses are intentional
  fragments (`...` placeholders, import+statement mixes) except `go-release/SKILL.md:157`,
  a go.mod `retract` block fenced as ```go.
- Go 1.26/1.27 API spot check against a real go1.27.3 toolchain (`GOTOOLCHAIN=go1.27.3 go doc`):
  errors.AsType, strings/bytes.CutLast, uuid, slog.NewMultiHandler, (*rand.Rand).N generic method,
  httptest.NewTestServer, T.ArtifactDir, synctest.Sleep, cryptotest.SetGlobalRandom all EXIST.
  runtime/secret and simd/archsimd exist but are build-constrained (GOEXPERIMENT), matching the skill's
  "non-production" framing.

## Environment notes
- Latest stable Go = go1.27.3 (from proxy.golang.org toolchain list). go.dev returns 403/000
  from this sandbox, so the workflow's `curl go.dev/dl` step cannot be reproduced here.
- Local go is 1.24.7; go1.27.3 fetched via GOTOOLCHAIN for verification.
- `git tag -l` is empty locally AND on origin. CHANGELOG links to releases/tag/v1.0.0 and v1.1.0.
