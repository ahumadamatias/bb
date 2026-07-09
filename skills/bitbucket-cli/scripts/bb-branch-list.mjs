#!/usr/bin/env node

import { spawnSync } from "node:child_process"

const args = process.argv.slice(2)
let limit = "30"
let workspace = ""
let repo = ""

for (let i = 0; i < args.length; i++) {
  const arg = args[i]
  if (arg === "--workspace") {
    workspace = args[++i] || ""
  } else if (arg === "--repo") {
    repo = args[++i] || ""
  } else if (!arg.startsWith("-") && limit === "30") {
    limit = arg
  } else {
    console.log(JSON.stringify({ ok: false, command: "bb branch list", error: `unknown argument: ${arg}` }, null, 2))
    process.exit(2)
  }
}

if (!/^\d+$/.test(limit)) {
  console.log(JSON.stringify({ ok: false, command: "bb branch list", error: `invalid limit: ${limit}` }, null, 2))
  process.exit(2)
}
if ((workspace && !repo) || (!workspace && repo)) {
  console.log(JSON.stringify({ ok: false, command: "bb branch list", error: "--workspace and --repo must be provided together" }, null, 2))
  process.exit(2)
}

console.error(`Listing branches with limit ${limit}${workspace ? ` for ${workspace}/${repo}` : ""}...`)

const bbArgs = ["branch", "list", "--limit", limit, "--output", "json"]
if (workspace) bbArgs.push("--workspace", workspace, "--repo", repo)

const result = spawnSync("bb", bbArgs, { encoding: "utf8" })
if (result.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb branch list", workspace: workspace || null, repo: repo || null, limit: Number(limit), error: `${result.stdout || ""}${result.stderr || ""}`.trim() }, null, 2))
  process.exit(result.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb branch list", workspace: workspace || null, repo: repo || null, limit: Number(limit), branches: JSON.parse(result.stdout) }, null, 2))
