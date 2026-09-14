# AutoLine Usage

## Scan a repository

```bash
autoline scan .
```

AutoLine reads repository manifests and lockfiles, reports the detected stack, and writes generated delivery assets.

## Replace generated assets

```bash
autoline scan . --force
```

Without `--force`, AutoLine refuses to overwrite an existing generated file. This protects hand-edited Dockerfiles and workflows.

## Generated assets

- `Dockerfile`: stack-aware multi-stage image build
- `.github/workflows/autoline.yml`: language-aware validation and Docker build
- `.pre-commit-config.yaml`: baseline repository hygiene hooks

Review generated files before production deployment, particularly for monorepos, custom application entrypoints, native dependencies, and organization-specific CI policies.
