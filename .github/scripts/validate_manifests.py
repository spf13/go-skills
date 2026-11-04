#!/usr/bin/env python3
"""Validate the catalog JSON against the published schemastore schemas.

Covers `.claude-plugin/marketplace.json` plus every `<skill>/.claude-plugin/plugin.json`.

The published plugin schema does not set `additionalProperties: false` at the root,
so documented fields the schema omits (notably `displayName`) validate cleanly. This
script deliberately does not add a stricter overlay: the published schema is the
contract, and local `claude plugin validate` in older CLI builds is stricter than it.
"""

from __future__ import annotations

import json
import sys
import urllib.request
from pathlib import Path

from jsonschema import Draft7Validator

from check_invariants import ROOT, declared_skills

MARKETPLACE_SCHEMA = "https://json.schemastore.org/claude-code-marketplace.json"
PLUGIN_SCHEMA = "https://json.schemastore.org/claude-code-plugin-manifest.json"


def fetch_schema(url: str) -> dict:
    with urllib.request.urlopen(url, timeout=30) as response:
        return json.load(response)


def validate(path: Path, schema: dict, schema_name: str, errors: list[str]) -> None:
    document = json.loads(path.read_text(encoding="utf-8"))
    validator = Draft7Validator(schema)
    for error in sorted(validator.iter_errors(document), key=lambda e: list(e.path)):
        where = "/".join(str(p) for p in error.absolute_path) or "<root>"
        errors.append(f"{path.relative_to(ROOT)} [{schema_name}] {where}: {error.message}")


def main() -> int:
    errors: list[str] = []

    marketplace_schema = fetch_schema(MARKETPLACE_SCHEMA)
    plugin_schema = fetch_schema(PLUGIN_SCHEMA)

    targets = [(ROOT / ".claude-plugin" / "marketplace.json", marketplace_schema, "marketplace")]
    targets += [(d / ".claude-plugin" / "plugin.json", plugin_schema, "plugin") for d in declared_skills().values()]

    for path, schema, schema_name in targets:
        validate(path, schema, schema_name, errors)

    if errors:
        print("manifest validation failed:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"validated {len(targets)} manifest(s) against the published schemastore schemas")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
