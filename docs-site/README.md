# AutoLine Documentation Site

This directory contains the static documentation website for AutoLine.

## GitHub Pages deployment

The site is deployed by the repository's GitHub Actions workflow to GitHub Pages. The intended public URL is:

`https://hunterkritik-byte.github.io/autoline-cli/`

The workflow publishes `docs-site/` as the Pages artifact. No Node.js build step is required; the site is static HTML/CSS and uses no external runtime dependencies.
