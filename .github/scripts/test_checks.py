#!/usr/bin/env python3
"""Tests for the skill-code check. Run: python3 -m unittest discover -s .github/scripts

Each test builds a small catalog in a temporary directory and runs the check's
main() against it, so the tests exercise the rules CONTRIBUTING.md documents
rather than helper internals. The checks need `go` on PATH and are skipped
without it.
"""

from __future__ import annotations

import contextlib
import io
import json
import re
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parent))

import check_skill_code  # noqa: E402

VERSION = "1.2.0"


def write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def skill_md(name: str, description: str = "Use whenever Go code is written, reviewed, or debugged.", body: str = "") -> str:
    return f"---\nname: {name}\ndescription: >\n  {description}\n---\n\n# {name}\n{body}"


class Catalog:
    """A minimal catalog on disk: one directory per skill, each a single-skill plugin."""

    def __init__(self, root: Path, names: tuple[str, ...] = ("alpha", "beta")) -> None:
        self.root = root
        for name in names:
            write(root / name / "SKILL.md", skill_md(name))
            write(root / name / ".claude-plugin" / "plugin.json", json.dumps({"name": name, "version": VERSION}))
        write(root / "CHANGELOG.md", f"# Changelog\n\n## [{VERSION}] - 2026-11-04\n")


def run(main) -> tuple[int, str]:
    """Run a check's main(), returning (exit code, stdout+stderr).

    Every check must exit non-zero exactly when it reports a problem, so CI can
    never go green over a failure it printed."""
    out = io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(out):
        try:
            code = main()
        except SystemExit as exit:
            code = exit.code if isinstance(exit.code, int) else 1
            out.write(str(exit.code))
    text = out.getvalue()
    reported = "failed (" in text or text.startswith("error:")
    if (code != 0) != reported:
        raise AssertionError(f"exit code {code} disagrees with the output:\n{text}")
    return code, text


class CatalogTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.tmp)
        self.catalog = Catalog(self.tmp)
        patch = mock.patch.object(check_skill_code, "ROOT", self.tmp)
        patch.start()
        self.addCleanup(patch.stop)


class DeclaredSkillsTest(CatalogTest):
    def test_only_plugin_directories_count(self) -> None:
        write(self.tmp / "templates" / "skill" / "SKILL.md", skill_md("templates"))
        write(self.tmp / "notes" / "SKILL.md", skill_md("notes"))
        found = check_skill_code.declared_skills()
        self.assertEqual(sorted(found), ["alpha", "beta"])


def go_minor() -> int:
    """Minor version of the `go` on PATH, or 0 without one."""
    if not shutil.which("go"):
        return 0
    out = subprocess.run(["go", "env", "GOVERSION"], capture_output=True, text=True).stdout
    match = re.search(r"go1\.(\d+)", out)
    return int(match.group(1)) if match else 0


@unittest.skipUnless(go_minor() >= 27, "needs Go 1.27 or newer on PATH")
class SkillCodeTest(CatalogTest):
    def check(self, body: str) -> tuple[int, str]:
        write(self.tmp / "alpha" / "SKILL.md", skill_md("alpha", body=body))
        return run(check_skill_code.main)

    def test_clean_skill_passes(self) -> None:
        code, out = self.check("Use `errors.Join` (Go 1.20).\n\n```go\nerr := errors.Join(a, b)\n```\n")
        self.assertEqual(code, 0, out)

    def test_fence_that_does_not_parse(self) -> None:
        self.assertIn("go fence does not parse", self.check("```go\nfunc (x {\n```\n")[1])

    def test_indented_fence_is_checked(self) -> None:
        self.assertIn("go fence does not parse", self.check("- item\n\n  ```go\n  func (x {\n  ```\n")[1])

    def test_fragments_with_elisions_parse(self) -> None:
        body = '```go\nimport "slices"\n\ntype T struct{}\n\nfunc (T) M() { ... }\n\nslices.Sort(s)\n...\n```\n'
        self.assertEqual(self.check(body)[0], 0)

    def test_invented_symbol_in_prose(self) -> None:
        self.assertIn("`errors.AsTypeOf[T]` is not in the standard library", self.check("Use `errors.AsTypeOf[T]`.\n")[1])

    def test_invented_selector_in_fence(self) -> None:
        self.assertIn("`strings.CutSuffixLast`", self.check("```go\nstrings.CutSuffixLast(s, \".\")\n```\n")[1])

    def test_non_std_module_names_are_skipped(self) -> None:
        body = "`runtime` here is `github.com/wailsapp/wails/v2/pkg/runtime`.\n\n```go\nruntime.OpenFileDialog(ctx, runtime.OpenDialogOptions{})\n```\n"
        self.assertEqual(self.check(body)[0], 0)

    def test_local_variable_named_like_a_package(self) -> None:
        self.assertEqual(self.check("```go\nurl := parse()\nfmt.Println(url.Host)\n```\n")[0], 0)

    def test_std_package_also_named_by_a_module_is_checked(self) -> None:
        body = 'Prefer `uuid` to `github.com/google/uuid`.\n\n```go\nimport "uuid"\n\nid := uuid.NewV9()\n```\n'
        self.assertIn("`uuid.NewV9`", self.check(body)[1])

    def test_range_over_package_call_is_checked(self) -> None:
        self.assertIn("`maps.Keyz`", self.check("```go\nfor k := range maps.Keyz(m) {\n    use(k)\n}\n```\n")[1])

    def test_file_names_are_not_symbols(self) -> None:
        self.assertEqual(self.check("Keep them in `errors.go` next to `doc.go`.\n")[0], 0)

    def test_declaration_in_one_fence_does_not_hide_another(self) -> None:
        body = "```go\nfunc New(fs Store) {}\n```\n\nThen:\n\n```go\nvar err *fs.PathErrorz\n```\n"
        self.assertIn("`fs.PathErrorz`", self.check(body)[1])

    def test_variadic_parameter_is_a_declaration(self) -> None:
        self.assertEqual(self.check("```go\nfunc join(path ...string) string {\n    return path.Cleanz\n}\n```\n")[0], 0)

    def test_var_and_for_bind_names(self) -> None:
        body = "```go\nvar url Endpoint\nuse(url.Hostz)\nfor _, path := range ps {\n    use(path.Basez)\n}\n```\n"
        self.assertEqual(self.check(body)[0], 0)

    def test_strings_and_comments_are_not_code(self) -> None:
        body = '```go\nmsg := "call strings.Bogus"\n// strings.Bogus is not real\nuse(msg)\n```\n'
        self.assertEqual(self.check(body)[0], 0)

    def test_test_receivers_resolve_to_testing(self) -> None:
        self.assertEqual(self.check("Use `b.Loop()` (Go 1.24) and `t.Context()`.\n")[0], 0)
        self.assertIn("`t.Contextz()` is not in the standard library", self.check("Use `t.Contextz()`.\n")[1])
        self.assertIn("`b.Loopz()` is not in the standard library", self.check("Use `b.Loopz()`.\n")[1])

    def test_module_names_in_prose_are_skipped(self) -> None:
        body = "`runtime` here is `github.com/wailsapp/wails/v2/pkg/runtime`; call `runtime.EventsEmit`.\n"
        self.assertEqual(self.check(body)[0], 0)
        self.assertIn("`runtime.EventsEmit`", self.check("Call `runtime.EventsEmit`.\n")[1])

    def test_std_import_overrides_module_name_in_prose(self) -> None:
        body = 'Not `github.com/google/uuid`: use `uuid.NewV9()`.\n\n```go\nimport "uuid"\n```\n'
        self.assertIn("`uuid.NewV9()` is not in the standard library", self.check(body)[1])

    def test_other_symbols_must_exist_by_the_claim(self) -> None:
        out = self.check("### `errors.Join` with `errors.AsType` (Go 1.20)\n")[1]
        self.assertIn("`errors.AsType` is used for Go 1.20 but arrived in Go 1.26", out)

    def test_failing_go_env_stops_clearly(self) -> None:
        failure = subprocess.CalledProcessError(1, ["go", "env"], stderr="go: download go1.99.0: toolchain not available\n")
        with mock.patch.object(check_skill_code.subprocess, "run", side_effect=failure):
            code, out = run(check_skill_code.main)
        self.assertEqual(code, 1)
        self.assertEqual(out, "error: `go env` failed: go: download go1.99.0: toolchain not available")

    def test_wrong_since_row(self) -> None:
        out = self.check("| Since | Feature |\n| --- | --- |\n| 1.22 | `slices` package |\n")[1]
        self.assertIn("`slices` is claimed for Go 1.22 but arrived in Go 1.21", out)

    def test_subject_after_context_package(self) -> None:
        out = self.check("| 1.26 | `math/rand/v2` `(*Rand).N` |\n")[1]
        self.assertIn("`(*Rand).N` is claimed for Go 1.26 but arrived in Go 1.27", out)

    def test_wrong_inline_claim(self) -> None:
        out = self.check("Use `sync.WaitGroup.Go` (Go 1.24) for this.\n")[1]
        self.assertIn("`sync.WaitGroup.Go` is claimed for Go 1.24 but arrived in Go 1.25", out)

    def test_non_subject_symbols_need_only_exist(self) -> None:
        self.assertEqual(self.check("### `errors.AsType[T]` over `errors.As` (Go 1.26)\n")[0], 0)

    def test_behavior_change_has_no_subject(self) -> None:
        self.assertEqual(self.check("| 1.26 | `B.Loop` no longer inhibits inlining |\n")[0], 0)

    def test_claim_newer_than_toolchain_stops_clearly(self) -> None:
        code, out = self.check("| 1.99 | `slices` package |\n")
        self.assertEqual(code, 1)
        self.assertIn("claims Go 1.99, newer than go1.", out)


if __name__ == "__main__":
    unittest.main()
