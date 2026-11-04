#!/usr/bin/env python3
"""Check the Go inside every SKILL.md against a real Go toolchain.

Three checks, all offline once a toolchain is installed:

1. Every ```go fence (indented ones included) parses with that toolchain's
   gofmt. Snippets may be fragments: a fence passes if it parses as a file, as
   declarations, as statements, or as declarations followed by statements.
   `{ ... }` and lines that are only `...` are treated as elided code.

2. Every standard-library symbol the skills name exists. That covers backticked
   prose (`errors.AsType`, `sync.WaitGroup.Go`, `t.Context()`) and `pkg.Name`
   selectors in fences, skipping names a fence imports from outside the standard
   library or declares itself.

3. Every standard-library version claim names the release that introduced it.
   A claim is a "| 1.X | ... |" row of a Since/Feature table, a heading ending in
   "(Go 1.X)", or a `symbol` written immediately before "(Go 1.X)". The claim's
   subject (its first std symbol, or a package named as "the X package") must
   have arrived exactly then; every other std token must merely exist by then.
   Claims that record a behavior change ("no longer") have no subject.

The oracle is the toolchain's stdversion manifest
($GOROOT/src/cmd/vendor/golang.org/x/tools/internal/stdlib/manifest.go), so the
toolchain must be at least as new as the newest release the skills claim; the
script says so and stops when it is not.
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

from check_invariants import declared_skills

MANIFEST = "src/cmd/vendor/golang.org/x/tools/internal/stdlib/manifest.go"
FENCE_RE = re.compile(r"^([ \t]*)```go\n(.*?)^\1```", re.S | re.M)
TICK_RE = re.compile(r"`([^`]+)`")
TABLE_ROW_RE = re.compile(r"^\|\s*1\.(\d+)\s*\|(.*)\|\s*$")
HEADING_CLAIM_RE = re.compile(r"^#+ .*\(Go 1\.(\d+)\+?\)\s*$")
INLINE_CLAIM_RE = re.compile(r"`([^`]+)`\s*\(Go 1\.(\d+)")
DECL_RE = re.compile(r"^(import|func|type|var|const)\b")
SELECTOR_RE = re.compile(r"(?<![\w.])([a-z]\w*)\.([A-Z]\w*)")
IMPORT_RE = re.compile(r'^\s*(?:import\s+)?(?:(\w+)\s+)?"([\w./-]+)"\s*$', re.M)
MODULE_PATH_RE = re.compile(r'[`"]((?:[\w-]+\.)+[\w-]+/[\w./-]+)[`"]')
TEST_RECEIVERS = {"t": "T", "b": "B", "f": "F", "T": "T", "B": "B", "F": "F"}

Manifest = dict[str, dict[str, int]]


def toolchain() -> tuple[Path, str]:
    """GOROOT and GOVERSION of the `go` on PATH."""
    try:
        out = subprocess.run(["go", "env", "GOROOT", "GOVERSION"], capture_output=True, text=True, check=True)
    except FileNotFoundError:
        raise SystemExit("error: `go` is not on PATH; install the latest stable Go")
    except subprocess.CalledProcessError as failure:
        raise SystemExit(f"error: `go env` failed: {failure.stderr.strip()}")
    root, version = out.stdout.splitlines()
    return Path(root), version


def load_manifest(path: Path) -> Manifest:
    """Package path -> symbol name -> minor version that introduced it."""
    manifest: Manifest = {}
    package: dict[str, int] | None = None
    for line in path.read_text(encoding="utf-8").splitlines():
        if match := re.match(r'^\t"([^"]+)": \{$', line):
            package = manifest.setdefault(match.group(1), {})
        elif package is not None and (match := re.match(r'^\t\t\{"([^"]+)", \w+, (\d+)[,}]', line)):
            package[match.group(1)] = int(match.group(2))
    return manifest


def fences(text: str) -> list[tuple[int, str]]:
    """(line, dedented source) of every ```go fence."""
    found = []
    for match in FENCE_RE.finditer(text):
        indent = match.group(1)
        lines = [line.removeprefix(indent) for line in match.group(2).splitlines()]
        found.append((text.count("\n", 0, match.start()) + 1, "\n".join(lines)))
    return found


def prose(text: str) -> list[tuple[int, str]]:
    """(line, text) of every line outside a fence."""
    found = []
    in_fence = False
    for number, line in enumerate(text.splitlines(), 1):
        if line.lstrip().startswith("```"):
            in_fence = not in_fence
        elif not in_fence:
            found.append((number, line))
    return found


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


# --- symbols -----------------------------------------------------------------


def packages_named(manifest: Manifest, name: str) -> list[str]:
    """Package paths a qualifier may mean: `slog` is log/slog, `rand` is math/rand or math/rand/v2."""
    if name in manifest:
        return [name]
    return [path for path in manifest if re.sub(r"/v\d+$", "", path).rsplit("/", 1)[-1] == name]


def resolve(manifest: Manifest, token: str, context: list[str]) -> list[int] | None:
    """Versions a std token could mean: [] if it names a std package but no such
    symbol, None if it is not a std reference at all."""
    token = re.sub(r"\[[^\]]*\]", "", token).strip()
    token = re.sub(r"\([^()]*\)$", "", token)  # call form: `slices.Sort(s)`, `t.Context()`
    if not re.fullmatch(r"[\w/.()*]+", token) or not token[-1:].isalnum():
        return None
    if token in manifest:
        return [min(manifest[token].values(), default=0)]

    head, _, rest = token.rpartition("/")
    pkg_part, _, name = rest.partition(".") if "." in rest else ("", "", rest)
    if head:
        pkg_part = f"{head}/{pkg_part}"
    if not name:
        return None

    if pkg_part in TEST_RECEIVERS and "." not in name:
        packages, name = ["testing"], f"{TEST_RECEIVERS[pkg_part]}.{name}"
    elif pkg_part and packages_named(manifest, pkg_part):
        if not name[:1].isupper():
            return None  # unexported, so not an API: `errors.go` is a file name
        packages = packages_named(manifest, pkg_part)
    elif context and (token.startswith("(") or pkg_part[:1].isupper()):
        packages, name = context, token  # Type.Method or (*Type).Method under a context package
    else:
        return None

    keys = [name]
    if match := re.fullmatch(r"(\w+)\.(\w+)", name):
        keys += [f"(*{match[1]}).{match[2]}", f"({match[1]}).{match[2]}"]
    return [manifest[p][k] for p in packages for k in keys if k in manifest[p]]


def declares(source: str, name: str) -> bool:
    """Whether code binds `name` itself (variable, parameter, range value)."""
    n = re.escape(name) + r"\b(?!\.)"
    return bool(
        re.search(rf"\b{n}\s*(,[\w\s,]*)?:?=[^=]", source)
        or re.search(rf"[(,]\s*{n}\s+(\.\.\.)?[\w*\[]", source)
        or re.search(rf"\b(var|for)\s+(\w+\s*,\s*)?{n}", source)
    )


def short_name(path: str) -> str:
    return re.sub(r"/v\d+$", "", path).rsplit("/", 1)[-1]


def imported_names(manifest: Manifest, source: str) -> tuple[set[str], set[str]]:
    """(std, non-std) package names a snippet imports."""
    std, other = set(), set()
    for alias, path in IMPORT_RE.findall(source):
        (std if path in manifest else other).add(alias or short_name(path))
    return std, other


def foreign_names(manifest: Manifest, text: str) -> set[str]:
    """Package names a SKILL.md takes from outside the standard library: non-std
    imports in its fences and module paths it mentions (Wails' `runtime`) — unless
    the skill also imports the std package of that name (`uuid`, `cmp`)."""
    std, other = imported_names(manifest, text)
    mentioned = {short_name(p) for p in MODULE_PATH_RE.findall(text) if p not in manifest}
    return (other | mentioned) - std


def code_only(source: str) -> str:
    return re.sub(r"//.*|\"(\\.|[^\"\n])*\"|`[^`]*`", "", source)


def unknown_selectors(manifest: Manifest, source: str, foreign: set[str]) -> list[str]:
    """`pkg.Name` selectors in a fence that name a std package but no std symbol.
    A name the fence declares itself is a variable, not a package; declarations in
    other fences do not count, so one `fs afero.Fs` parameter cannot hide `fs.*` typos."""
    std, other = imported_names(manifest, source)
    foreign = (foreign | other) - std
    code = code_only(source)
    unknown = []
    for qualifier, name in SELECTOR_RE.findall(code):
        packages = packages_named(manifest, qualifier)
        if not packages or qualifier in foreign or declares(code, qualifier):
            continue
        if not any(name in manifest[p] for p in packages) and f"{qualifier}.{name}" not in unknown:
            unknown.append(f"{qualifier}.{name}")
    return unknown


# --- version claims ----------------------------------------------------------


def claims(lines: list[tuple[int, str]]) -> list[tuple[int, int, list[str], str]]:
    """(line, minor, tokens, claim text) for every version claim in prose lines."""
    found: list[tuple[int, int, list[str], str]] = []
    for number, line in lines:
        if match := TABLE_ROW_RE.match(line):
            found.append((number, int(match.group(1)), TICK_RE.findall(match.group(2)), match.group(2)))
        elif match := HEADING_CLAIM_RE.match(line):
            found.append((number, int(match.group(1)), TICK_RE.findall(line), line))
        else:
            for match in INLINE_CLAIM_RE.finditer(line):
                found.append((number, int(match.group(2)), [match.group(1)], match.group(0)))
    return found


def check_claim(manifest: Manifest, minor: int, tokens: list[str], text: str) -> tuple[int, list[str]]:
    """Return (checked token count, problems) for one claim."""
    context = [t for t in tokens if t in manifest]
    resolved = [(t, v) for t in tokens if (v := resolve(manifest, t, context))]
    subject = None
    if "no longer" not in text:
        symbols = [t for t, _ in resolved if t not in manifest]
        if symbols:
            subject = symbols[0]
        elif resolved and ("package" in text or text.strip(" `|") == resolved[0][0]):
            subject = resolved[0][0]

    problems = []
    for token, versions in resolved:
        if token == subject and minor not in versions:
            problems.append(f"`{token}` is claimed for Go 1.{minor} but arrived in Go 1.{min(versions)}")
        elif token != subject and min(versions) > minor:
            problems.append(f"`{token}` is used for Go 1.{minor} but arrived in Go 1.{min(versions)}")
    return len(resolved), problems


def main() -> int:
    root, version = toolchain()
    manifest_path = root / MANIFEST
    if not manifest_path.is_file():
        raise SystemExit(f"error: {version} has no stdversion manifest at {manifest_path}; use the latest stable Go")
    manifest = load_manifest(manifest_path)
    current = int(re.search(r"go1\.(\d+)", version).group(1))
    errors: list[str] = []
    counts = {"fences": 0, "symbols": 0, "claims": 0}

    for name, directory in declared_skills().items():
        text = (directory / "SKILL.md").read_text(encoding="utf-8")
        where = f"{name}/SKILL.md"
        foreign = foreign_names(manifest, text)
        for line, source in fences(text):
            counts["fences"] += 1
            if (error := parse_error(root / "bin" / "gofmt", source)) is not None:
                errors.append(f"{where}:{line}: go fence does not parse: {error}")
            errors += [f"{where}:{line}: go fence uses `{s}`, which is not in the standard library" for s in unknown_selectors(manifest, source, foreign)]

        lines = prose(text)
        for line, content in lines:
            tokens = TICK_RE.findall(content)
            context = [t for t in tokens if t in manifest]
            for token in tokens:
                if token.split(".", 1)[0] in foreign:
                    continue
                versions = resolve(manifest, token, context)
                if versions == []:
                    errors.append(f"{where}:{line}: `{token}` is not in the standard library")
                elif versions:
                    counts["symbols"] += 1

        for line, minor, tokens, claim in claims(lines):
            if minor > current:
                raise SystemExit(f"error: {where}:{line} claims Go 1.{minor}, newer than {version}; use the latest stable Go")
            checked, problems = check_claim(manifest, minor, tokens, claim)
            counts["claims"] += checked
            errors += [f"{where}:{line}: {p}" for p in problems]

    if errors:
        print(f"skill code check failed ({len(errors)} problem(s)):", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(
        f"skill code holds on {version}: {counts['fences']} go fence(s) parse, "
        f"{counts['symbols']} std symbol mention(s) exist, {counts['claims']} version claim token(s) match"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
