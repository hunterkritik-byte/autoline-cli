#!/usr/bin/env python3
"""Optional Python-powered repository insights for AutoLine.

Uses only the standard library and never executes project code.
"""
from __future__ import annotations
import argparse
import json
import os
from collections import Counter
from pathlib import Path

IGNORED = {".git", "node_modules", "vendor", "target", ".venv", "venv", "dist", "build", ".cache", ".gradle"}
EXTENSIONS = {
    ".py": "Python", ".js": "JavaScript", ".jsx": "JavaScript", ".ts": "TypeScript", ".tsx": "TypeScript",
    ".go": "Go", ".rs": "Rust", ".java": "Java", ".kt": "Kotlin", ".kts": "Kotlin", ".cs": "C#",
    ".php": "PHP", ".rb": "Ruby", ".ex": "Elixir", ".exs": "Elixir", ".dart": "Dart", ".swift": "Swift",
    ".c": "C", ".h": "C/C++", ".cc": "C++", ".cpp": "C++", ".cxx": "C++", ".scala": "Scala",
    ".hs": "Haskell", ".lua": "Lua", ".r": "R", ".jl": "Julia", ".zig": "Zig", ".pl": "Perl",
    ".sh": "Shell", ".bash": "Shell",
}
MANIFESTS = {
    "package.json": "Node.js", "requirements.txt": "Python", "pyproject.toml": "Python", "go.mod": "Go",
    "Cargo.toml": "Rust", "pom.xml": "Java", "build.gradle": "Kotlin/Gradle", "build.gradle.kts": "Kotlin/Gradle",
    "composer.json": "PHP", "Gemfile": "Ruby", "mix.exs": "Elixir", "pubspec.yaml": "Dart/Flutter",
    "Package.swift": "Swift", "CMakeLists.txt": "C/C++", "build.sbt": "Scala", "stack.yaml": "Haskell",
    "Makefile": "Native/Make", "flake.nix": "Nix", "deno.json": "Deno",
}

def iter_files(root: Path):
    for base, dirs, files in os.walk(root):
        dirs[:] = [d for d in dirs if d not in IGNORED and not d.startswith(".")]
        for name in files:
            path = Path(base) / name
            try:
                if path.is_file() and path.stat().st_size <= 2_000_000:
                    yield path
            except OSError:
                continue

def analyze(root: Path) -> dict:
    languages, manifests = Counter(), Counter()
    files = lines = todos = credential_signals = 0
    for path in iter_files(root):
        files += 1
        if path.name in MANIFESTS:
            manifests[MANIFESTS[path.name]] += 1
        language = EXTENSIONS.get(path.suffix.lower())
        if language:
            languages[language] += 1
        try:
            text = path.read_text(encoding="utf-8", errors="ignore")
        except OSError:
            continue
        lines += text.count("\n") + (1 if text else 0)
        todos += sum(text.count(marker) for marker in ("TODO", "FIXME"))
        lowered = text.lower()
        credential_signals += sum(lowered.count(marker) for marker in ("api_key=", "secret_key=", "private_key="))
    return {
        "root": str(root), "files": files, "lines": lines,
        "languages": dict(languages.most_common()), "manifests": dict(manifests.most_common()),
        "todo_fixme_markers": todos, "credential_pattern_signals": credential_signals,
        "note": "Heuristic signals only; credential_pattern_signals never exposes matched values.",
    }

def main() -> int:
    parser = argparse.ArgumentParser(description="Generate lightweight repository insights")
    parser.add_argument("path", nargs="?", default=".")
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args()
    result = analyze(Path(args.path).resolve())
    if args.json:
        print(json.dumps(result, indent=2, sort_keys=True))
    else:
        print(f"Repository: {result['root']}\nFiles: {result['files']}\nLines: {result['lines']}")
        print("Languages:")
        for language, count in result["languages"].items(): print(f"  {language:<16} {count}")
        print(f"TODO/FIXME: {result['todo_fixme_markers']}")
        print(f"Credential-pattern signals: {result['credential_pattern_signals']}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
