#!/usr/bin/env node

import { spawnSync } from "node:child_process"

const id = process.argv[2]
const strategy = process.argv[3]?.startsWith("--") ? "merge-commit" : process.argv[3] || "merge-commit"
const closeSourceBranch = process.argv.includes("--close-source-branch")
const allowed = new Set(["merge-commit", "squash", "fast-forward"])

if (!id || !/^\d+$/.test(id)) {
  console.log(JSON.stringify({ ok: false, command: "bb pr merge", error: "usage: node scripts/bb-pr-merge.mjs <id> [strategy] [--close-source-branch]" }, null, 2))
  process.exit(2)
}
if (!allowed.has(strategy)) {
  console.log(JSON.stringify({ ok: false, command: "bb pr merge", id: Number(id), error: `invalid strategy: ${strategy}` }, null, 2))
  process.exit(2)
}

console.error(`Merging pull request #${id} with ${strategy}...`)

const args = ["pr", "merge", id, "--strategy", strategy]
if (closeSourceBranch) args.push("--close-source-branch")

const result = spawnSync("bb", args, { encoding: "utf8" })
if (result.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb pr merge", id: Number(id), strategy, close_source_branch: closeSourceBranch, error: `${result.stdout || ""}${result.stderr || ""}`.trim() }, null, 2))
  process.exit(result.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb pr merge", id: Number(id), strategy, close_source_branch: closeSourceBranch, output: result.stdout.trim() }, null, 2))
