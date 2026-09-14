# AutoLine

> ⚡ High-performance, zero-config repository automation for modern engineering teams.

[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/) [![CI](https://github.com/hunterkritik-byte/autoline-cli/actions/workflows/test.yml/badge.svg)](https://github.com/hunterkritik-byte/autoline-cli/actions/workflows/test.yml)

## 💖 Sponsor This Project

**Help fund faster, cheaper cloud-native delivery.** AutoLine automates multi-stage Docker builds, dependency caching, CI validation, and developer-quality hooks. Better Docker layer reuse can reduce repeated CI compute and build time for teams with frequent container builds, with potential savings that scale with workload.

If AutoLine saves your team engineering time or CI spend, please consider sponsoring development. Corporate sponsorship supports new detectors, safer generators, benchmarks, and enterprise-ready CI/CD features.

**Sponsorship & partnership inquiries:** hunterkritik@gmail.com

## What it does

AutoLine scans a repository and generates delivery assets without executing project commands during scanning.

- Detects Node.js, Python, Go, Rust, and generic projects
- Detects npm, pnpm, Yarn, pip, Poetry, uv, Go modules, and Cargo signals
- Generates multi-stage Dockerfiles with BuildKit dependency caching
- Generates GitHub Actions workflows with language-aware caching
- Generates pre-commit hooks for common repository hygiene checks
- Protects existing generated files unless `--force` is supplied
- Keeps detection and generation modular for easy extension

## Quick start

```bash
go install github.com/hunterkritik-byte/autoline-cli/cmd/autoline@latest
cd your-project
autoline scan .
```

To intentionally replace existing generated assets:

```bash
autoline scan . --force
```

Generated files:

```text
Dockerfile
.github/workflows/autoline.yml
.pre-commit-config.yaml
```

## Supported stacks

| Stack | Detection | Docker strategy | CI setup |
| --- | --- | --- | --- |
| Node.js | package.json + lockfile | dependency-layer caching | setup-node |
| Python | requirements.txt / pyproject.toml | pip cache + multi-stage | setup-python |
| Go | go.mod | module/build cache + distroless runtime | setup-go |
| Rust | Cargo.toml | Cargo registry/target cache | Rust toolchain |
| Generic | fallback | reviewable Alpine base | Docker build |

## Architecture

```text
cmd/autoline/main.go          CLI and command UX
internal/detector/            manifest and lockfile detection
internal/generator/           Docker, CI, and pre-commit generation
docs/design.md                architecture and safety model
.github/workflows/test.yml    AutoLine project validation
```

## Development

```bash
go mod tidy
go test ./...
go vet ./...
go run ./cmd/autoline scan .
```

Generated assets are templates, not deployment guarantees. Review custom build outputs, native dependencies, monorepo layouts, secrets, and deployment-specific requirements before production use.

## Roadmap

- Monorepo/workspace-aware generation
- Framework-specific build artifact detection
- Registry-backed BuildKit cache configuration
- Additional CI providers
- Container image metadata, SBOM, and provenance options
- Snapshot tests for generated assets

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Focused pull requests with tests are welcome.

## License

MIT. See `LICENSE` when distributed with a release.
