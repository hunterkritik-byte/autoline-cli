# Python Tooling

AutoLine keeps the primary CLI in Go for portability and uses small optional Python tools for repository intelligence.

## Repository insights

```bash
python3 tools/autoline_insights.py .
python3 tools/autoline_insights.py . --json
make insights
```

The analyzer reports language/file counts, common manifests, TODO/FIXME markers, and credential-pattern signals. It never executes project code and never prints matched credential values.

## Design principle

Python is an optional specialist layer rather than a runtime dependency of the Go CLI. This keeps the default AutoLine binary easy to distribute while allowing richer analyzers to evolve independently.

Future specialist tools can use the same JSON-oriented contract for dependency graphs, framework detection, documentation analysis, and benchmark/report generation.
