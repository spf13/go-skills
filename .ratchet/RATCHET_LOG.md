# Ratchet log

Process note: implementer passes run in the main session (it holds the inventory context); every judge
is a fresh subagent that never wrote the change.

### S1 · fidelity — KEPT — 68→80 — 3fe48f0
- Critique: 88 Go fences and ~80 std version claims had no mechanical check (BASELINE); go-release/SKILL.md:157 fenced go.mod as ```go.
- Change: add .github/scripts/check_skill_code.py (gofmt parse of every go fence; std version claims vs the toolchain's stdversion manifest), a `skill code` CI job on setup-go stable, CONTRIBUTING rule; refence go.mod as text.
- Verified: `GOTOOLCHAIN=go1.27.3 python3 .github/scripts/check_skill_code.py` → 87 fences, 80 claims OK; invariants, manifests, go_release go1.27.3, `claude plugin validate` ×7 green. Mutation: 7/8 caught (implementer), 16/22 version + 5/5 fence (judge).
- Judge: KEEP — real, regression-free; follow-ups: nonexistent symbols skipped, context-package subject gap, indented fences skipped, tracebacks on missing go/manifest, old-manifest false errors, CONTRIBUTING wording, no Unreleased entry, no unit tests.

### S1 · correctness — KEPT — 70→86 — 825fdb8
- Critique: cobra-viper/SKILL.md:384-405, :486, go/SKILL.md:794, go-spec-reviewer/SKILL.md:59 claimed Go 1.27 encoding/json rejects duplicate names and invalid UTF-8; false on go1.27.3 (v1 API keeps v1 semantics; only encoding/json/v2 is strict — src/encoding/json/v2_options.go:89-96). go/SKILL.md:914 and go-release/SKILL.md:114 forecast nogreenteagc removal in 1.27; it still builds on 1.27.3.
- Change: rewrite the four passages; cobra-viper shows validating config bytes with encoding/json/v2 before Viper (compiled + run vs viper v1.21.0); add .github/claims stdlib test module (json v1 lenient / v2 strict, unbuffered timers) + CI step; CHANGELOG [Unreleased].
- Verified: failing test of old claim → FAIL on go1.27.3; claims suite ok; check_skill_code 88/80, invariants, go_release, `claude plugin validate` green.
- Judge: KEEP — old claim reproduced false, replacement verified true. Minor: nojsonv2 breaks the json/v2 snippet (unmentioned); timer test doesn't pin asynctimerchan removal; json.Options mention dropped.
