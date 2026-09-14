# BuildKit and CI Caching

AutoLine places dependency and compiler caches on BuildKit cache mounts so rebuilds can reuse downloaded artifacts without baking cache contents into the final image.

## Docker cache mounts

Node.js uses `/root/.npm` for npm, `/root/.local/share/pnpm/store` for pnpm, and `/usr/local/share/.cache/yarn` for Yarn. Go uses `/go/pkg/mod` for modules and `/root/.cache/go-build` for compiled packages. Rust uses `/usr/local/cargo/registry` and the build target directory. Python uses `/root/.cache/pip`.

These mounts are ephemeral BuildKit cache state: they accelerate rebuilds but are not copied into the runtime stage.

## CI caching

GitHub Actions uses the native cache modes provided by `setup-node` and `setup-go`, while Docker Buildx performs the image build. GitLab and Bitbucket pipelines use their provider-native Docker execution paths and keep the image/SBOM workflow explicit.

The cost model is straightforward: cache hits reduce repeated package downloads and compiler work, shortening CI wall time and lowering network and compute consumption. Cache effectiveness depends on runner retention and the provider's cache policy.

## Correctness

Cache mounts must never be treated as a source of application state. A lockfile remains the dependency source of truth, and immutable install modes are preferred for pnpm, Yarn, Cargo, and uv where supported by the generated command.

When optimizing a generated pipeline, measure cache hit rate, build duration, and cache storage cost rather than assuming every cache is beneficial.
