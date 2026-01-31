# API Gateway Documentation

This directory contains the VitePress documentation site for API Gateway.

## Development

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run docs:dev
```

The documentation site will be available at `http://localhost:5173/` (note: in development mode, base path is `/`, not `/api-gateway/`).

## Build

Build the documentation site:

```bash
npm run docs:build
```

The built site will be in `.vitepress/dist`.

## Preview

Preview the built site:

```bash
npm run docs:preview
```

## Deployment

The documentation is automatically deployed to GitHub Pages via GitHub Actions when changes are pushed to the `master` branch.

See `.github/workflows/docs.yml` for the deployment configuration.

## Structure

- `index.md` - Homepage (English)
- `zh/index.md` - Homepage (Chinese)
- `guide/` - Guide documentation
- `api/` - API reference
- `examples/` - Example configurations
- `.vitepress/` - VitePress configuration
