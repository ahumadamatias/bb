---
name: bitbucket-cli
description: Use when an AI agent needs to operate Bitbucket Cloud through the `bb` CLI: check auth, inspect repos/branches/PRs, review diffs, comment, approve, or merge pull requests.
---

# Bitbucket CLI

Use the local `bb` CLI to inspect and operate Bitbucket Cloud from an AI agent.
Prefer JSON-producing scripts for context-efficient, machine-readable results.

## How It Works

1. Check that `bb` is installed and authenticated with `bb-check.mjs`.
2. Use read-only scripts first: list PRs, view PR details, and fetch diffs.
3. Present findings to the user before taking mutating actions.
4. Only comment, approve, or merge when the user explicitly asks for that action.
5. Use `references/command-reference.md` only when you need raw command details.

## Safety Rules

- Never run `bb pr merge` unless the user explicitly asks to merge a specific PR.
- Never run `bb pr approve` unless the user explicitly asks to approve or remove approval.
- Before commenting inline, verify the target `path` and `line` are from the PR diff.
- If `bb auth status` fails, ask the user to run `bb auth login`; do not request tokens in chat.
- Prefer `--output json` for read operations.

## Usage

```bash
node scripts/bb-check.mjs
node scripts/bb-pr-list.mjs [state] [limit]
node scripts/bb-pr-view.mjs <id> [--diff]
node scripts/bb-pr-comment.mjs <id> <body> [path] [line] [--task]
node scripts/bb-pr-merge.mjs <id> [strategy] [--close-source-branch]
```

Arguments:

`bb-pr-list.mjs`:

- `state` - `OPEN`, `MERGED`, or `DECLINED` (defaults to `OPEN`)
- `limit` - max PRs to return (defaults to `30`; `0` means all)

`bb-pr-view.mjs`:

- `id` - pull request ID
- `--diff` - include the PR diff

`bb-pr-comment.mjs`:

- `id` - pull request ID
- `body` - comment text
- `path` - optional file path for inline comments
- `line` - optional line in the new file version; requires `path`
- `--task` - also create a blocking Bitbucket task

`bb-pr-merge.mjs`:

- `id` - pull request ID
- `strategy` - `merge-commit`, `squash`, or `fast-forward` (defaults to `merge-commit`)
- `--close-source-branch` - close the source branch after merge

Examples:

```bash
node scripts/bb-check.mjs
node scripts/bb-pr-list.mjs OPEN 20
node scripts/bb-pr-view.mjs 42 --diff
node scripts/bb-pr-comment.mjs 42 "Please fix this" internal/api/client.go 42 --task
node scripts/bb-pr-merge.mjs 42 squash --close-source-branch
```

## Output

Scripts write status messages to stderr and JSON to stdout.

Example `bb-pr-list.mjs` output:

```json
{
  "ok": true,
  "command": "bb pr list",
  "state": "OPEN",
  "limit": 30,
  "pull_requests": []
}
```

Example failure output:

```json
{
  "ok": false,
  "command": "bb auth status",
  "error": "No hay credenciales configuradas. Ejecutá 'bb auth login' o seteá BB_EMAIL y BB_TOKEN."
}
```

## Present Results to User

Use this format for PR reviews:

```markdown
**PR #<id>: <title>**
State: <state>
Source -> Dest: <source> -> <dest>

Findings:
- <severity>: <file:line> <issue>

Suggested Actions:
- <comment/approve/merge recommendation>
```

If no issues are found, say that explicitly and mention any residual risks.

## Troubleshooting

- `bb` not found: install it with `go install github.com/ahumadamatias/bb/cmd/bb@latest` and ensure `$HOME/go/bin` is in `PATH`.
- Auth fails: run `bb auth login`, then retry `node scripts/bb-check.mjs`.
- Missing scopes: recreate the Atlassian API token with the Bitbucket scopes documented in the project README.
- Repo inference fails: run the command inside a Bitbucket clone with `origin` pointing to bitbucket.org, or pass `--workspace` and `--repo` manually with raw `bb`.
- Stale Go module proxy during install: use `GOPROXY=direct go install github.com/ahumadamatias/bb/cmd/bb@latest`.

## End-User Installation

For public distribution with `skills.sh`:

```bash
npx skills add ahumadamatias/bb --skill bitbucket-cli
```
