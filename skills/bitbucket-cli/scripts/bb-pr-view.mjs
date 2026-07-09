#!/usr/bin/env node

import { spawnSync } from "node:child_process"

const id = process.argv[2]
const includeDiff = process.argv.includes("--diff")

if (!id || !/^\d+$/.test(id)) {
  console.log(JSON.stringify({ ok: false, command: "bb pr view", error: "usage: node scripts/bb-pr-view.mjs <id> [--diff]" }, null, 2))
  process.exit(2)
}

console.error(`Viewing pull request #${id}${includeDiff ? " with diff" : ""}...`)

const args = ["pr", "view", id, "--output", "json"]
if (includeDiff) args.splice(3, 0, "--diff")

const result = spawnSync("bb", args, { encoding: "utf8" })
if (result.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb pr view", id: Number(id), error: `${result.stdout || ""}${result.stderr || ""}`.trim() }, null, 2))
  process.exit(result.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb pr view", id: Number(id), pull_request: JSON.parse(result.stdout) }, null, 2))
