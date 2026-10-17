---
name: conventional-commit
description: >-
  Generate Conventional Commit messages with a domain scope and a changelog body for this project,
  splitting by domain when needed, and always confirming with the user before committing.
  Trigger when the user asks to commit changes, write a commit message, or perform git commits.
---

# Conventional Commit Workflow

When asked to commit changes in this repository, follow this exact procedure.

## 1. Format

```text
<type>(<domain>): <short summary>

- <changelog bullet 1>
- <changelog bullet 2>
```

**Types** (pick the one that best matches the change):
* `feat` — new feature / domain module / endpoint / capability
* `fix` — bug fix
* `chore` — tooling, dependencies, config, scripts, non-production code
* `refactor` — code change that neither fixes a bug nor adds a feature
* `docs` — documentation only (including `.agents/**`, `AGENTS.md`, `README.md`)
* `test` — adding or fixing unit/integration tests
* `perf` — performance or query optimization
* `style` — formatting, whitespace only, no behavior change
* `build` — build system, Go toolchain, Makefile, Dockerfile
* `ci` — CI/CD pipeline changes

**Domain** = the module or area affected:
- Modules in `internal/modules/`: `auth`, `user`, `product`, `order`, etc.
- Core areas: `bootstrap`, `config`, `middleware`, `pkg`, `scripts`, `agents`, `deps`.

**Summary**: imperative mood, lowercase, no trailing period (e.g. `feat(auth): implement refresh token rotation`, not `feat(auth): Implemented refresh token rotation.`).

**Changelog body**: bullet list detailing specific changes made, one line per meaningful change.

Example:
```text
feat(auth): add logout endpoint and refresh token cookie handling

- Add /auth/logout endpoint to invalidate refresh token cookie
- Set HttpOnly, Secure, and SameSite attributes on refresh cookie
- Update auth delivery DTO and handler to clear cookie on logout
```

## 2. One Domain Per Commit

Before drafting a commit message, inspect the diff (`git status`, `git diff`). If the changes span **more than one domain** (e.g. touched both `internal/modules/user` and `internal/modules/auth`, or mixed code changes with `.agents/**` doc updates), **split into multiple commits** — one per domain — instead of bundling everything into one commit.
- Stage only the files relevant to each domain (`git add <specific files>`), never a blanket `git add .` across unrelated domains.

## 3. Always Confirm Before Committing

Never run `git commit` without asking first. For each proposed commit:

1. Show the exact commit message and list of staged files to the user.
2. Ask for explicit approval (e.g., *"Apakah pesan commit ini sudah sesuai? [y/n]"*).
3. Only run `git commit` after explicit approval. If rejected, ask what to adjust.

## 4. Standard Git Safety & Verification

- Always run tests and verify build (`go test ./...` and `go build ./cmd/api`) before proposing a commit.
- Never commit secrets, `.env` files, or binary outputs (`bin/`).
- Create new commits rather than `--amend`ing published ones.
