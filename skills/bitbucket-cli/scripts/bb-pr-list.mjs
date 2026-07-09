#!/usr/bin/env node

import { spawnSync } from "node:child_process"

const state = (process.argv[2] || "OPEN").toUpperCase()
const limit = process.argv[3] || "30"
const allowed = new Set(["OPEN", "MERGED", "DECLINED"])

if (!allowed.has(state)) {
  console.log(JSON.stringify({ ok: false, command: "bb pr list", error: `invalid state: ${state}` }, null, 2))
  process.exit(2)
}

console.error(`Listing ${state} pull requests...`)

const result = spawnSync("bb", ["pr", "list", "--state", state, "--limit", limit, "--output", "json"], { encoding: "utf8" })
if (result.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb pr list", state, limit: Number(limit), error: `${result.stdout || ""}${result.stderr || ""}`.trim() }, null, 2))
  process.exit(result.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb pr list", state, limit: Number(limit), pull_requests: JSON.parse(result.stdout) }, null, 2))
