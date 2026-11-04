# Ratchet log

Process note: implementer passes run in the main session (it holds the inventory context); every judge
is a fresh subagent that never wrote the change.

### S1 · fidelity — KEPT — 68→80 — 3fe48f0
- Critique: 88 Go fences and ~80 std version claims had no mechanical check (BASELINE); go-release/SKILL.md:157 fenced go.mod as ```go.
- Change: add .github/scripts/check_skill_code.py (gofmt parse of every go fence; std version claims vs the toolchain's stdversion manifest), a `skill code` CI job on setup-go stable, CONTRIBUTING rule; refence go.mod as text.
- Verified: `GOTOOLCHAIN=go1.27.3 python3 .github/scripts/check_skill_code.py` → 87 fences, 80 claims OK; invariants, manifests, go_release go1.27.3, `claude plugin validate` ×7 green. Mutation: 7/8 caught (implementer), 16/22 version + 5/5 fence (judge).
- Judge: KEEP — real, regression-free; follow-ups: nonexistent symbols skipped, context-package subject gap, indented fences skipped, tracebacks on missing go/manifest, old-manifest false errors, CONTRIBUTING wording, no Unreleased entry, no unit tests.
