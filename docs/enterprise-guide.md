# AutoLine Enterprise Guide

AutoLine is intentionally split into detection, planning, generation, and CI validation so teams can adopt it without coupling repository discovery to one deployment platform.

## Recommended workflow

```text
Repository
    │
    ▼
DetectWorkspace
    │
    ├── application manifests
    ├── lockfiles/package managers
    └── independent monorepo modules
    │
    ▼
Provider plan
    │
    ├── GitHub Actions
    ├── GitLab CI
    └── Bitbucket Pipelines
    │
    ▼
Safe generation
    │
    ├── Dockerfile
    ├── CI pipeline
    ├── pre-commit
    ├── OCI metadata
    └── SBOM automation
```

## CLI contract

### Scan

```bash
autoline scan . --dry-run --json
autoline scan . --provider=github
autoline scan . --provider=gitlab
autoline scan . --provider=bitbucket
autoline scan . --force
```

`--dry-run` performs detection and planning without writing files. `--json` is designed for scripts and CI orchestration. Existing generated assets are protected unless `--force` is explicitly supplied.

### Doctor

Use the environment diagnostic before relying on generated assets:

```bash
autoline doctor
autoline doctor --json
```

The command checks Git, Docker, Docker Buildx, and pre-commit. A failed check produces a non-zero exit status, which makes it suitable for CI gates.

## Extension model

New detectors should remain side-effect free and return a `Stack`. Provider-specific output belongs in the generator layer. This keeps detection reusable for future providers and allows snapshot tests to validate generated assets independently.

## Safety principles

- Never execute repository build or package-manager commands during detection.
- Do not overwrite existing generated files without `--force`.
- Keep generated CI permissions minimal.
- Treat generated Dockerfiles and pipelines as reviewable source, not deployment guarantees.
- Run tests, `go vet`, and the race detector before publishing changes.

## Testing strategy

AutoLine uses unit tests plus immutable template snapshots. CI validates the module graph, runs `go test -race ./...`, runs `go vet ./...`, builds the CLI, and exercises the help/diagnostic surfaces.

## Production rollout

1. Run `autoline scan . --dry-run --json` and review the plan.
2. Commit generated assets through normal code review.
3. Pin action/tool versions according to your organization's supply-chain policy.
4. Configure registry-backed BuildKit caches where appropriate.
5. Review secrets, native dependencies, runtime users, and deployment-specific requirements.
6. Re-run the scanner after major repository layout changes.
