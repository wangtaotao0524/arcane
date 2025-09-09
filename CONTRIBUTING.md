# Contributing to Arcane

Thanks for helping make Arcane better! Contributions of all kinds are welcome.

## Ways to Contribute

- Reporting bugs: use the GitHub templates — Bug Report and Feature Request.
- Suggesting features or improvements.
- Code contributions (frontend or backend).

## Quick Start

Prereqs:

- Go 1.25 and Docker
- Node 24 and pnpm 10.15

1. Fork and clone:

   ```bash
   git clone https://github.com/<your-username>/arcane.git
   cd arcane
   ```

2. Create a branch:

   ```bash
   git switch -c feat/my-new-feature
   # or
   git switch -c fix/issue-123
   ```

3. Frontend (SvelteKit):

   ```bash
   cd frontend
   pnpm install
   pnpm dev
   ```

4. Backend (Go + Gin):

   ```bash
   go install github.com/air-verse/air@latest
   cd backend
   air
   ```

## Code Style

- Conventional Commits for messages:

  ```bash
  git add .
  git commit -m "feat: add X"
  # or
  git commit -m "fix: resolve Y"
  ```

- Frontend lint/format:

  ```bash
  cd frontend
  pnpm check
  pnpm format
  ```

- Backend basics:

  ```bash
  cd backend
  go fmt ./...
  go vet ./...
  ```

## Pull Requests

- Keep changes focused and small when possible.
- Include context in the PR description and link issues (e.g., “Closes #123”).
- Be ready to iterate based on review feedback.

## Issues & Feedback

- Bug Report: https://github.com/ofkm/arcane/issues/new?template=bug.yml
- Feature Request: https://github.com/ofkm/arcane/issues/new?template=feature.yml
