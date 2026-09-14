# AutoLine Architecture

AutoLine separates repository understanding from artifact generation. The detector produces structured signals; the generation engine consumes those signals without needing to understand every source-language detail.

## Detection pipeline

1. **Filesystem walk** discovers manifests and lockfiles while skipping `.git`, dependency trees, build output, caches, and virtual environments.
2. **Signal extraction** identifies language, package manager, dependency files, build command, runtime command, and binary metadata.
3. **Workspace model** groups independent modules into a deterministic `Workspace` ordered by relative path.
4. **Generation** maps each stack to Docker and CI templates.

The current implementation intentionally uses manifest signals instead of a full language AST. This keeps scans fast and makes detection deterministic across CI and developer machines.

## Generation engine

The generator is provider-aware and exposes stable primitives:

- `Files` keeps GitHub generation backwards compatible.
- `FilesForProvider` selects GitHub Actions, GitLab CI, or Bitbucket Pipelines.
- `GenerateWorkspace` applies the same safety rules to every detected module.
- Existing files are protected unless `--force` is explicitly supplied.

## Enterprise boundaries

Generated Dockerfiles use multi-stage builds, BuildKit caches, slim or distroless runtimes, OCI labels, and CI-driven SBOM generation. Generated CI configurations request read-only repository contents permissions where the provider supports it.

AutoLine does not silently modify application source code, credentials, deployment targets, or production configuration. Generated assets remain reviewable text files.

## Extension model

New stacks should add a detector signal set and a generator template rather than coupling detection to a particular CI provider. New providers should implement their pipeline template and file name while reusing the stack-specific build logic.
