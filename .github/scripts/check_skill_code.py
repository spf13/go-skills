#!/usr/bin/env python3
"""Check the Go inside every SKILL.md against a real Go toolchain.

Two checks, both offline once a toolchain is installed:

1. Every ```go fence parses with that toolchain's gofmt. Snippets may be
   fragments: a fence passes if it parses as a file, as declarations, as
   statements, or as declarations followed by statements. `{ ... }` and lines
   that are only `...` are treated as elided code.

2. Every standard-library version claim names the release that introduced the
   symbol, as recorded in the toolchain's own stdversion manifest
   ($GOROOT/src/cmd/vendor/golang.org/x/tools/internal/stdlib/manifest.go).
   A claim is a "| 1.X | ... |" row of a Since/Feature table, a heading ending
   in "(Go 1.X)", or a `symbol` written immediately before "(Go 1.X)".
   Language features and tokens that are not std symbols are skipped.

Use a toolchain at least as new as the newest release the skills cover
(CI installs the latest stable Go); older toolchains lack the newer symbols.
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

from check_invariants import ROOT, declared_skills

MANIFEST = "src/cmd/vendor/golang.org/x/tools/internal/stdlib/manifest.go"
FENCE_RE = re.compile(r"^```go\n(.*?)^```", re.S | re.M)
TICK_RE = re.compile(r"`([^`]+)`")
TABLE_ROW_RE = re.compile(r"^\|\s*1\.(\d+)\s*\|(.*)\|\s*$")
HEADING_CLAIM_RE = re.compile(r"^#+ .*\(Go 1\.(\d+)\+?\)\s*$")
INLINE_CLAIM_RE = re.compile(r"`([^`]+)`\s*\(Go 1\.(\d+)")
DECL_RE = re.compile(r"^(import|func|type|var|const)\b")
TEST_RECEIVERS = {"t": "T", "b": "B", "f": "F", "T": "T", "B": "B", "F": "F"}

Manifest = dict[str, dict[str, int]]


def goroot() -> Path:
    out = subprocess.run(["go", "env", "GOROOT"], capture_output=True, text=True, check=True)
    return Path(out.stdout.strip())


def load_manifest(path: Path) -> Manifest:
    """Package path -> symbol name -> minor version that introduced it."""
    manifest: Manifest = {}
    package: dict[str, int] | None = None
    for line in path.read_text(encoding="utf-8").splitlines():
        if match := re.match(r'^\t"([^"]+)": \{$', line):
            package = manifest.setdefault(match.group(1), {})
        elif package is not None and (match := re.match(r'^\t\t\{"([^"]+)", \w+, (\d+),', line)):
            package[match.group(1)] = int(match.group(2))
    return manifest


# --- fence parsing -----------------------------------------------------------


def split_decls(source: str) -> tuple[str, str]:
    """Split a snippet into top-level declarations and loose statements."""
    decls: list[str] = []
    stmts: list[str] = []
    in_decl = False
    for line in source.splitlines():
        if not in_decl and DECL_RE.match(line):
            decls.append(line)
            in_decl = line.rstrip().endswith(("{", "("))
        elif in_decl:
            decls.append(line)
            if line.startswith(("}", ")")):
                in_decl = False
        else:
            stmts.append(line)
    return "\n".join(decls), "\n".join(stmts)


def candidates(source: str) -> list[str]:
    source = re.sub(r"\{\s*\.\.\.\s*\}", "{}", source)
    source = re.sub(r"^\s*\.\.\.\s*$", "", source, flags=re.M)
    decls, stmts = split_decls(source)
    return [
        source,
        f"package p\n{source}",
        f"package p\nfunc _() {{\n{source}\n}}\n",
        f"package p\n{decls}\nfunc _() {{\n{stmts}\n}}\n",
    ]


def parse_error(gofmt: Path, source: str) -> str | None:
    """None if any candidate wrapping parses; else the error for the plain snippet."""
    first: str | None = None
    for candidate in candidates(source):
        result = subprocess.run([str(gofmt), "-e"], input=candidate, capture_output=True, text=True)
        if result.returncode == 0:
            return None
        if first is None:
            first = result.stderr.strip().splitlines()[0] if result.stderr.strip() else "parse error"
    return first


# --- version claims ----------------------------------------------------------


def packages_named(manifest: Manifest, name: str) -> list[str]:
    """Package paths a qualifier may mean: `slog` is log/slog, `rand` is math/rand or math/rand/v2."""
    if name in manifest:
        return [name]
    return [path for path in manifest if re.sub(r"/v\d+$", "", path).rsplit("/", 1)[-1] == name]


def resolve(manifest: Manifest, token: str, context: list[str]) -> list[int] | None:
    """Versions a token could mean, or None if it is not a std symbol or package."""
    token = re.sub(r"\[[^\]]*\]", "", token).removesuffix("()").strip()
    if not re.fullmatch(r"[\w/.()*]+", token):
        return None
    if token in manifest:
        return [min(manifest[token].values(), default=0)]

    head, _, rest = token.rpartition("/")
    pkg_part, _, name = (rest.partition(".") if "." in rest else ("", "", rest))
    if head:
        pkg_part = f"{head}/{pkg_part}"
    if not name:
        return None

    if pkg_part in TEST_RECEIVERS and "." not in name:
        packages, name = ["testing"], f"{TEST_RECEIVERS[pkg_part]}.{name}"
    elif pkg_part and packages_named(manifest, pkg_part):
        packages = packages_named(manifest, pkg_part)
    elif token.startswith("(") or pkg_part[:1].isupper():
        packages, name = context, token  # Type.Method or (*Type).Method under a context package
    else:
        return None

    keys = [name]
    if m := re.fullmatch(r"(\w+)\.(\w+)", name):
        keys = [f"(*{m[1]}).{m[2]}", f"({m[1]}).{m[2]}"]
    found = [manifest[p][k] for p in packages for k in keys if k in manifest.get(p, {})]
    return found or None


def claims(text: str) -> list[tuple[int, int, list[str], str]]:
    """(line, minor, tokens, claim text) for every version claim in a SKILL.md body."""
    found: list[tuple[int, int, list[str], str]] = []
    in_fence = False
    for number, line in enumerate(text.splitlines(), 1):
        if line.startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if match := TABLE_ROW_RE.match(line):
            found.append((number, int(match.group(1)), TICK_RE.findall(match.group(2)), match.group(2)))
        elif match := HEADING_CLAIM_RE.match(line):
            found.append((number, int(match.group(1)), TICK_RE.findall(line), line))
        else:
            for match in INLINE_CLAIM_RE.finditer(line):
                found.append((number, int(match.group(2)), [match.group(1)], match.group(0)))
    return found


def check_claim(manifest: Manifest, minor: int, tokens: list[str], text: str) -> tuple[int, list[str]]:
    """Return (checked token count, problems) for one claim.

    Every std token must exist by the claimed release. The claim's subject (its
    first std token) must have arrived exactly then, unless it is a package named
    only as context ("`net/url` Clones") or the claim records a behavior change
    ("`B.Loop` no longer inhibits inlining").
    """
    context = [t for t in tokens if t in manifest]
    problems: list[str] = []
    checked = 0
    for token in tokens:
        versions = resolve(manifest, token, context)
        if versions is None:
            continue
        checked += 1
        subject = checked == 1 and "no longer" not in text
        if token in manifest:
            subject = subject and ("package" in text or text.strip(" `|") == token)
        if subject and minor not in versions:
            problems.append(f"`{token}` is claimed for Go 1.{minor} but arrived in Go 1.{min(versions)}")
        elif not subject and min(versions) > minor:
            problems.append(f"`{token}` is used for Go 1.{minor} but arrived in Go 1.{min(versions)}")
    return checked, problems


def main() -> int:
    root = goroot()
    manifest = load_manifest(root / MANIFEST)
    gofmt = root / "bin" / "gofmt"
    errors: list[str] = []
    fences = claimed = 0

    for name, directory in declared_skills().items():
        path = directory / "SKILL.md"
        text = path.read_text(encoding="utf-8")
        for match in FENCE_RE.finditer(text):
            fences += 1
            if (error := parse_error(gofmt, match.group(1))) is not None:
                line = text.count("\n", 0, match.start()) + 1
                errors.append(f"{name}/SKILL.md:{line}: go fence does not parse: {error}")
        for line, minor, tokens, claim in claims(text):
            checked, problems = check_claim(manifest, minor, tokens, claim)
            claimed += checked
            errors += [f"{name}/SKILL.md:{line}: {p}" for p in problems]

    if errors:
        print(f"skill code check failed ({len(errors)} problem(s)):", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    version = subprocess.run(["go", "env", "GOVERSION"], capture_output=True, text=True).stdout.strip()
    print(f"skill code holds: {fences} go fence(s) parse, {claimed} std version claim(s) match {version}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
