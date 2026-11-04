#!/usr/bin/env python3
"""Tests for the CI checks. Run: python3 -m unittest discover -s .github/scripts

Each test builds a small catalog in a temporary directory and runs a check's
main() against it, so the tests exercise the rules CONTRIBUTING.md documents
rather than helper internals. The skill-code tests need `go` on PATH and are
skipped without it.
"""

from __future__ import annotations

import contextlib
import io
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parent))

import check_go_release  # noqa: E402
import check_invariants  # noqa: E402
import check_skill_code  # noqa: E402

VERSION = "1.2.0"


def write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def skill_md(name: str, description: str = "Use whenever Go code is written, reviewed, or debugged.", body: str = "") -> str:
    return f"---\nname: {name}\ndescription: >\n  {description}\n---\n\n# {name}\n{body}"


class Catalog:
    """A minimal valid catalog on disk; tests break one rule at a time."""

    def __init__(self, root: Path, names: tuple[str, ...] = ("alpha", "beta")) -> None:
        self.root = root
        self.entries = []
        for name in names:
            self.add_skill(name)
        self.write_catalog()
        write(root / "CHANGELOG.md", f"# Changelog\n\n## [{VERSION}] - 2026-11-04\n")

    def add_skill(self, name: str, keywords: list[str] | None = None, version: str = VERSION) -> None:
        keywords = keywords or ["go", name]
        write(self.root / name / "SKILL.md", skill_md(name))
        write(
            self.root / name / ".claude-plugin" / "plugin.json",
            json.dumps({"name": name, "version": version, "keywords": keywords}),
        )
        self.entries.append({"name": name, "source": f"./{name}", "category": "development", "tags": list(keywords)})

    def write_catalog(self) -> None:
        write(self.root / ".claude-plugin" / "marketplace.json", json.dumps({"plugins": self.entries}))
        rows = "".join(f"| [`{e['name']}`](./{e['name']}/SKILL.md) | x | x |\n" for e in self.entries)
        write(self.root / "README.md", f"| Skill | What | Install |\n| --- | --- | --- |\n{rows}")


def run(main) -> tuple[int, str]:
    """Run a check's main(), returning (exit code, stdout+stderr)."""
    out = io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(out):
        try:
            code = main()
        except SystemExit as exit:
            code = exit.code if isinstance(exit.code, int) else 1
            out.write(str(exit.code))
    return code, out.getvalue()


class CatalogTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.tmp)
        self.catalog = Catalog(self.tmp)
        # An empty repository: no tags, no HEAD, so the tag rules see a clean slate
        # however the machine running the tests is set up.
        subprocess.run(["git", "init", "-q", str(self.tmp)], check=True)
        for patch in (
            mock.patch.object(check_invariants, "ROOT", self.tmp),
            mock.patch.dict(os.environ, {"GITHUB_REF_TYPE": "", "GITHUB_REF_NAME": ""}),
        ):
            patch.start()
            self.addCleanup(patch.stop)


class FrontmatterTest(unittest.TestCase):
    def parse(self, text: str) -> dict[str, str]:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "SKILL.md"
            path.write_text(text, encoding="utf-8")
            return check_invariants.frontmatter(path)

    def test_folded_scalar_joins_lines(self) -> None:
        fields = self.parse("---\nname: go\ndescription: >\n  first line,\n  current through Go 1.27.\n---\nbody\n")
        self.assertEqual(fields, {"name": "go", "description": "first line, current through Go 1.27."})

    def test_plain_value_keeps_colons(self) -> None:
        self.assertEqual(self.parse("---\ndescription: Use it: always\n---\n")["description"], "Use it: always")

    def test_no_frontmatter(self) -> None:
        self.assertEqual(self.parse("# just a heading\n"), {})

    def test_stops_at_closing_fence(self) -> None:
        self.assertNotIn("later", self.parse("---\nname: go\n---\nlater: value\n"))


class InvariantsTest(CatalogTest):
    def check(self) -> tuple[int, str]:
        return run(check_invariants.main)

    def test_valid_catalog_passes(self) -> None:
        code, out = self.check()
        self.assertEqual(code, 0, out)
        self.assertIn(f"version {VERSION}", out)

    def test_orphan_skill_directory(self) -> None:
        write(self.tmp / "gamma" / "SKILL.md", skill_md("gamma"))
        write(self.tmp / "gamma" / ".claude-plugin" / "plugin.json", json.dumps({"name": "gamma", "version": VERSION}))
        code, out = self.check()
        self.assertEqual(code, 1)
        self.assertIn("skill directory 'gamma' has no matching marketplace entry", out)

    def test_entry_with_extra_key(self) -> None:
        self.catalog.entries[0]["description"] = "duplicated"
        self.catalog.write_catalog()
        code, out = self.check()
        self.assertEqual(code, 1)
        self.assertIn("entry 'alpha': keys must be exactly", out)

    def test_unknown_category(self) -> None:
        self.catalog.entries[0]["category"] = "tools"
        self.catalog.write_catalog()
        self.assertIn("category must be one of", self.check()[1])

    def test_tags_must_equal_keywords(self) -> None:
        self.catalog.entries[1]["tags"] = ["go"]
        self.catalog.write_catalog()
        self.assertIn("'beta': entry tags ['go'] != plugin.json keywords", self.check()[1])

    def test_frontmatter_name_must_match_directory(self) -> None:
        write(self.tmp / "alpha" / "SKILL.md", skill_md("not-alpha"))
        self.assertIn("SKILL.md frontmatter name is 'not-alpha'", self.check()[1])

    def test_short_description(self) -> None:
        write(self.tmp / "alpha" / "SKILL.md", skill_md("alpha", description="Go."))
        self.assertIn("description is 3 chars; minimum is 40", self.check()[1])

    def test_versions_must_agree(self) -> None:
        self.catalog.add_skill("gamma", version="1.3.0")
        self.catalog.write_catalog()
        self.assertIn("plugin versions must be identical", self.check()[1])

    def test_changelog_needs_an_entry_for_the_version(self) -> None:
        write(self.tmp / "CHANGELOG.md", "# Changelog\n\n## [1.1.0]\n")
        self.assertIn(f"CHANGELOG.md has no '## [{VERSION}]' entry", self.check()[1])

    def test_readme_table_must_list_every_skill(self) -> None:
        write(self.tmp / "README.md", "| [`alpha`](./alpha/SKILL.md) | x | x |\n")
        self.assertIn("README skills table must list every skill exactly once", self.check()[1])

    def test_tagged_commit_must_carry_its_version(self) -> None:
        with mock.patch.dict(os.environ, {"GITHUB_REF_TYPE": "tag", "GITHUB_REF_NAME": "v9.9.9"}):
            self.assertIn(f"HEAD is tagged v9.9.9 but the catalog version is {VERSION}", self.check()[1])

    def test_version_behind_latest_tag(self) -> None:
        with mock.patch.object(check_invariants, "latest_version_tag", return_value="2.0.0"):
            self.assertIn(f"catalog version {VERSION} is older than the latest tag v2.0.0", self.check()[1])


class GoReleaseTest(CatalogTest):
    def setUp(self) -> None:
        super().setUp()
        claim = skill_md("alpha", description="Idiomatic Go, current through Go 1.27. Use whenever Go code is written.")
        write(self.tmp / "alpha" / "SKILL.md", claim)

    def check(self, latest: str) -> tuple[int, str]:
        with mock.patch.object(sys, "argv", ["check_go_release.py", latest]):
            return run(check_go_release.main)

    def test_claim_matches_latest(self) -> None:
        self.assertEqual(self.check("go1.27.3")[0], 0)

    def test_claim_behind_latest(self) -> None:
        code, out = self.check("go1.28.0")
        self.assertEqual(code, 1)
        self.assertIn("claims 'current through Go 1.27' but the latest stable Go release is go1.28.0", out)
        self.assertIn("behind", out)

    def test_known_claiming_skill_must_keep_its_claim(self) -> None:
        with mock.patch.object(check_go_release, "CLAIMING_SKILLS", frozenset({"beta"})):
            self.assertIn("'beta': SKILL.md description lost its", self.check("go1.27.3")[1])

    def test_unparseable_latest(self) -> None:
        self.assertEqual(self.check("tip"), (1, "error: cannot parse a Go version from 'tip'\n"))


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
        self.assertEqual(self.check("```go\nfunc join(path ...string) string {\n    return path.Clean\n}\n```\n")[0], 0)

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
