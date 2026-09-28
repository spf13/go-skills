#!/usr/bin/env python3
"""Assert each skill's "current through Go X.Y" claim matches the latest stable Go.

Rule documented in CONTRIBUTING.md ("Go release currency"): a skill whose
frontmatter `description` carries a `current through Go X.Y` claim must name the
latest stable Go release. The claim moves only in the same change that adds the
release's features, so a shipped release turns the scheduled check red.

The latest stable version is supplied by the caller, never fetched here: the
scheduled workflow reads it from https://go.dev/dl/?mode=json and passes it as
the first argument (or as GO_RELEASE_LATEST). Supplying it keeps the comparison
logic exercisable offline and pins the red/green demonstration to a fixed value.
"""

from __future__ import annotations

import os
import re
import sys
from pathlib import Path

from check_invariants import EXCLUDED, ROOT, frontmatter

# Skills known to carry a `current through Go X.Y` claim. This is the recorded
# baseline for "the claim must not silently disappear": the repository has no
# other persistent record, so the list lives with the check. A skill absent from
# this set may still make a claim — it is then checked like any other.
CLAIMING_SKILLS = frozenset({"go"})

CLAIM_RE = re.compile(r"current through Go (\d+\.\d+)")
MINOR_RE = re.compile(r"(\d+)\.(\d+)")

errors: list[str] = []


def fail(message: str) -> None:
    errors.append(message)


def skill_dirs() -> dict[str, Path]:
    """Root-level directories that are single-skill plugins."""
    found: dict[str, Path] = {}
    for path in sorted(ROOT.iterdir()):
        if not path.is_dir() or path.name in EXCLUDED:
            continue
        if (path / ".claude-plugin" / "plugin.json").is_file():
            found[path.name] = path
    return found


def requested_latest() -> str | None:
    """The latest stable version handed over by the workflow, argument first."""
    if len(sys.argv) > 1 and sys.argv[1].strip():
        return sys.argv[1].strip()
    return os.environ.get("GO_RELEASE_LATEST", "").strip() or None


def minor(version: str) -> tuple[int, int] | None:
    """The (major, minor) release line of a Go version such as 'go1.27.3' or '1.27'."""
    match = MINOR_RE.search(version)
    if match is None:
        return None
    return (int(match.group(1)), int(match.group(2)))


def main() -> int:
    latest = requested_latest()
    if latest is None:
        print(
            "error: no latest Go version supplied; pass it as the first argument "
            "or set GO_RELEASE_LATEST",
            file=sys.stderr,
        )
        return 1

    latest_minor = minor(latest)
    if latest_minor is None:
        print(f"error: cannot parse a Go version from {latest!r}", file=sys.stderr)
        return 1

    checked = 0
    for name, directory in skill_dirs().items():
        description = frontmatter(directory / "SKILL.md").get("description", "")
        match = CLAIM_RE.search(description)
        if match is None:
            if name in CLAIMING_SKILLS:
                fail(
                    f"'{name}': SKILL.md description lost its 'current through Go X.Y' claim; "
                    f"{name} is a known claiming skill, so the claim must not disappear"
                )
            continue

        claimed = match.group(1)
        claimed_minor = minor(claimed)
        if claimed_minor is None:
            fail(f"'{name}': cannot parse the claimed Go line from 'current through Go {claimed}'")
            continue
        checked += 1

        if claimed_minor != latest_minor:
            relation = "behind" if claimed_minor < latest_minor else "ahead of"
            fail(
                f"'{name}': SKILL.md claims 'current through Go {claimed}' but the latest "
                f"stable Go release is {latest} — the claim is {relation} the release"
            )

    if errors:
        print(f"go release check failed ({len(errors)} problem(s)):", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"go release claims hold: {checked} claim(s) match the latest stable release {latest}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
