#!/usr/bin/env python3
"""Assert the cross-file invariants documented in CONTRIBUTING.md.

Every rule in CONTRIBUTING.md maps to a check here. If you add a rule there,
add the check here too — and vice versa.
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TAXONOMY = ("languages", "development", "frameworks")
ENTRY_KEYS = {"name", "source", "category", "tags"}
# Directories that are never skills, besides hidden ones (.git, .github, tool
# state). Anything without a plugin.json is skipped anyway.
EXCLUDED = {"templates"}
MIN_DESCRIPTION = 40

def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def frontmatter(path: Path) -> dict[str, str]:
    """Parse the leading YAML frontmatter, handling folded (>) and literal (|) scalars."""
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0].strip() != "---":
        return {}
    fields: dict[str, str] = {}
    key: str | None = None
    for line in lines[1:]:
        if line.strip() == "---":
            break
        if line[:1].strip() and ":" in line:
            key, _, value = line.partition(":")
            key = key.strip()
            value = value.strip()
            fields[key] = "" if value in (">", "|", ">-", "|-", ">+", "|+") else value
        elif key and line[:1] in (" ", "\t"):
            fields[key] = f"{fields[key]} {line.strip()}".strip()
    return fields


def declared_skills() -> dict[str, Path]:
    """Root-level directories that are single-skill plugins."""
    found: dict[str, Path] = {}
    for path in sorted(ROOT.iterdir()):
        if not path.is_dir() or path.name in EXCLUDED or path.name.startswith("."):
            continue
        if (path / ".claude-plugin" / "plugin.json").is_file():
            found[path.name] = path
    return found


def readme_table_skills(fail) -> list[str]:
    """Skill names linked from the skills table in README.md."""
    pattern = re.compile(r"^\|\s*\[`([^`]+)`\]\(\./([^/]+)/SKILL\.md\)")
    names: list[str] = []
    for line in (ROOT / "README.md").read_text(encoding="utf-8").splitlines():
        match = pattern.match(line)
        if match:
            if match.group(1) != match.group(2):
                fail(f"README table row '{match.group(1)}' links to ./{match.group(2)}/SKILL.md")
            names.append(match.group(1))
    return names


def latest_version_tag() -> str | None:
    result = subprocess.run(
        ["git", "tag", "-l", "v*", "--sort=-v:refname"],
        cwd=ROOT, capture_output=True, text=True, check=False,
    )
    for tag in result.stdout.split():
        if re.fullmatch(r"v\d+\.\d+\.\d+", tag):
            return tag[1:]
    return None


def tag_at_head() -> str | None:
    """Version of a `v*` tag pointing at HEAD, in CI or locally."""
    if os.environ.get("GITHUB_REF_TYPE") == "tag":
        name = os.environ.get("GITHUB_REF_NAME", "")
        return name[1:] if re.fullmatch(r"v\d+\.\d+\.\d+", name) else None
    result = subprocess.run(
        ["git", "tag", "--points-at", "HEAD", "-l", "v*"],
        cwd=ROOT, capture_output=True, text=True, check=False,
    )
    for tag in result.stdout.split():
        if re.fullmatch(r"v\d+\.\d+\.\d+", tag):
            return tag[1:]
    return None


def semver(value: str) -> tuple[int, int, int] | None:
    match = re.fullmatch(r"(\d+)\.(\d+)\.(\d+)", value)
    if match is None:
        return None
    return (int(match.group(1)), int(match.group(2)), int(match.group(3)))


def main() -> int:
    errors: list[str] = []
    fail = errors.append
    marketplace = read_json(ROOT / ".claude-plugin" / "marketplace.json")
    entries = marketplace.get("plugins", [])
    skills = declared_skills()

    entry_names = [entry.get("name") for entry in entries]
    if len(entry_names) != len(set(entry_names)):
        fail("marketplace.json has duplicate entry names")
    for name in sorted(set(entry_names) - set(skills)):
        fail(f"marketplace entry '{name}' has no matching skill directory")
    for name in sorted(set(skills) - set(entry_names)):
        fail(f"skill directory '{name}' has no matching marketplace entry")

    table_names = readme_table_skills(fail)
    if sorted(table_names) != sorted(entry_names):
        fail(
            "README skills table must list every skill exactly once: "
            f"table has {sorted(table_names)}, catalog has {sorted(entry_names)}"
        )

    versions: dict[str, str] = {}

    for entry in entries:
        name = entry.get("name")
        if set(entry) != ENTRY_KEYS:
            fail(f"entry '{name}': keys must be exactly {sorted(ENTRY_KEYS)}, got {sorted(entry)}")
        if entry.get("source") != f"./{name}":
            fail(f"entry '{name}': source must be './{name}', got {entry.get('source')!r}")
        if entry.get("category") not in TAXONOMY:
            fail(f"entry '{name}': category must be one of {list(TAXONOMY)}, got {entry.get('category')!r}")

        directory = skills.get(name)
        if directory is None:
            continue

        manifest = read_json(directory / ".claude-plugin" / "plugin.json")
        skill = frontmatter(directory / "SKILL.md")

        if manifest.get("name") != name:
            fail(f"'{name}': plugin.json name is {manifest.get('name')!r}; must equal the directory name")
        if skill.get("name") != name:
            fail(f"'{name}': SKILL.md frontmatter name is {skill.get('name')!r}; must equal the directory name")

        description = skill.get("description", "")
        if not description:
            fail(f"'{name}': SKILL.md frontmatter is missing a description")
        elif len(description) < MIN_DESCRIPTION:
            fail(f"'{name}': SKILL.md description is {len(description)} chars; minimum is {MIN_DESCRIPTION}")

        tags, keywords = entry.get("tags", []), manifest.get("keywords", [])
        if set(tags) != set(keywords):
            fail(f"'{name}': entry tags {sorted(tags)} != plugin.json keywords {sorted(keywords)}")
        elif len(tags) != len(set(tags)) or len(keywords) != len(set(keywords)):
            fail(f"'{name}': tags and keywords must each be free of duplicates")

        versions[name] = manifest.get("version", "")

    distinct = sorted(set(versions.values()))
    if len(distinct) > 1:
        fail(f"plugin versions must be identical across the catalog, got {versions}")
    elif distinct:
        version = distinct[0]
        released = tag_at_head()
        latest = latest_version_tag()
        current, previous = semver(version), semver(latest) if latest else None
        if released is not None and released != version:
            fail(f"HEAD is tagged v{released} but the catalog version is {version}")
        if current is not None and previous is not None and current < previous:
            fail(f"catalog version {version} is older than the latest tag v{latest}")
        changelog = ROOT / "CHANGELOG.md"
        if not changelog.is_file():
            fail("CHANGELOG.md is missing")
        elif not re.search(rf"^##\s*\[{re.escape(version)}\]", changelog.read_text(encoding="utf-8"), re.M):
            fail(f"CHANGELOG.md has no '## [{version}]' entry")

    if errors:
        print(f"invariant check failed ({len(errors)} problem(s)):", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"invariants hold for {len(skills)} skill(s); version {distinct[0] if distinct else 'n/a'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
