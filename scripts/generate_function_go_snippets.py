from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOC_DIR = ROOT / "snippets" / "functions"


def extract_signature(text: str) -> str:
    match = re.search(r"- Signature: `([^`]+)`", text)
    if not match:
        raise ValueError("missing signature")
    return match.group(1).strip()


def extract_source(text: str) -> str:
    match = re.search(r"- Source: `([^`]+)`", text)
    if not match:
        raise ValueError("missing source")
    return match.group(1).strip()


def extract_summary(text: str) -> str:
    marker = "## What This Function Needs To Do"
    idx = text.find(marker)
    if idx == -1:
        return "Implement the behavior described in the sibling markdown specification."
    rest = text[idx + len(marker):].strip().splitlines()
    for line in rest:
        line = line.strip()
        if not line or line.startswith("-") or line.startswith("## "):
            continue
        return line
    return "Implement the behavior described in the sibling markdown specification."


def make_go_stub(source: str, signature: str, summary: str, stem: str) -> str:
    return "\n".join(
        [
            "//go:build ignore",
            "",
            "package snippets",
            "",
            f"// Source: {source}",
            f"// Spec: {stem}.md",
            f"// {summary}",
            signature + " {",
            f'\tpanic("TODO: implement according to {stem}.md")',
            "}",
            "",
        ]
    )


def main() -> None:
    generated = 0
    for md_path in sorted(DOC_DIR.glob("*.md")):
        if md_path.name == "README.md":
            continue
        text = md_path.read_text(encoding="utf-8")
        source = extract_source(text)
        signature = extract_signature(text)
        summary = extract_summary(text)
        go_path = md_path.with_suffix(".go")
        go_path.write_text(make_go_stub(source, signature, summary, md_path.stem), encoding="utf-8")
        generated += 1
    print(f"Generated {generated} Go function snippets in {DOC_DIR}")


if __name__ == "__main__":
    main()
