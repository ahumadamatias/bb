# Command Reference

This reference summarizes the `bb` commands most useful to an AI agent.

## Authentication

```bash
bb auth status
bb auth login
```

`bb auth login` stores credentials in `~/.config/bb/config.yaml`, or
`$XDG_CONFIG_HOME/bb/config.yaml` when `XDG_CONFIG_HOME` is set.

## Read Operations

```bash
bb repo list --output json
bb branch list --output json
bb pr list --state OPEN --limit 30 --output json
bb pr view 42 --output json
bb pr view 42 --diff --output json
```

Use `--workspace` and `--repo` when not running inside a Bitbucket git clone:

```bash
bb pr list --workspace acme --repo api --output json
```

## PR Comments

General comment:

```bash
bb pr comment 42 --body "LGTM"
```

Inline comment:

```bash
bb pr comment 42 --body "Please fix this" --path internal/api/client.go --line 42
```

Blocking task:

```bash
bb pr comment 42 --body "Please fix this before merge" --path internal/api/client.go --line 42 --task
```

## Approval

```bash
bb pr approve 42
bb pr approve 42 --remove
```

## Merge

```bash
bb pr merge 42
bb pr merge 42 --strategy squash --close-source-branch
```

Supported strategies: `merge-commit`, `squash`, `fast-forward`.

## Global Flags

```bash
--email EMAIL
--token TOKEN
--workspace WORKSPACE
--repo REPO
--output json
```

Precedence: flags > environment variables > config file.

Environment variables:

```bash
BB_EMAIL
BB_TOKEN
BB_WORKSPACE
```
