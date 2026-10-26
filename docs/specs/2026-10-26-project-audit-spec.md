# Living Specification: `go-skills` Catalog Overhaul (`2026-10-26`)

- **Source Audit**: `docs/audits/2026-10-26-project-audit.md` (`R-01` – `R-07`)
- **Existing Kata Epics**: `j6sd`, `6c4t`, `d96p`, `g6k7`, `zr5c`
- **Status**: `READY FOR EXECUTION`

---

## 1. Core Invariants
1. **Mutually Exclusive Trigger Boundaries**: Every `*/SKILL.md` and `*/.claude-plugin/plugin.json` description MUST include both positive trigger contexts and explicit negative boundaries (`Not for...`).
2. **Progressive Disclosure Budget (≤500 Lines)**: Every `*/SKILL.md` MUST stay under 500 lines; detailed version tables and API references live in `<skill>/references/*.md`.
3. **Zero Unresolved Placeholders & 100% Valid Go Samples**: No `SKILL.md` may contain unresolved `[ALL_CAPS]` placeholders outside code fences, and every Go code sample must use valid stdlib/library symbols (`logger.With`, not `slog.With`).
4. **Single Rule Ownership Across Skills**: Each architectural rule (`cmd/` layout, `slog` DI, `os.Root` vs `pathologize.Join`, module-floor version gating) has one canonical owner and is cross-referenced—never contradicted or re-derived—by sibling skills.
