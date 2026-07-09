#!/usr/bin/env node

import { spawnSync } from "node:child_process"
import { mkdtempSync, rmSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"

const args = process.argv.slice(2)
let branch = ""
let filePath = ""
let workspace = ""
let repo = ""
let remote = ""
let tempDir = ""

function cleanup() {
  if (tempDir) rmSync(tempDir, { recursive: true, force: true })
}

process.on("exit", cleanup)
process.on("SIGINT", () => {
  cleanup()
  process.exit(130)
})
process.on("SIGTERM", () => {
  cleanup()
  process.exit(143)
})

function fail(error, status = 1, extra = {}) {
  console.log(JSON.stringify({ ok: false, command: "git branch code", branch: branch || null, ...extra, error }, null, 2))
  process.exit(status)
}

function output(result) {
  return `${result.stdout || ""}${result.stderr || ""}`.trim()
}

for (let i = 0; i < args.length; i++) {
  const arg = args[i]
  if (arg === "--workspace") {
    workspace = args[++i] || ""
  } else if (arg === "--repo") {
    repo = args[++i] || ""
  } else if (arg === "--remote") {
    remote = args[++i] || ""
  } else if (!arg.startsWith("-") && !branch) {
    branch = arg
  } else if (!arg.startsWith("-") && !filePath) {
    filePath = arg
  } else {
    fail(`unknown argument: ${arg}`, 2)
  }
}

if (!branch) {
  fail("usage: node scripts/bb-branch-code.mjs <branch> [path] [--workspace ws --repo repo | --remote url]", 2)
}
if ((workspace && !repo) || (!workspace && repo)) {
  fail("--workspace and --repo must be provided together", 2)
}
if (remote && (workspace || repo)) {
  fail("use either --remote or --workspace/--repo, not both", 2)
}

let gitDir = ""
let ref = `origin/${branch}`

if (workspace) remote = `https://bitbucket.org/${workspace}/${repo}.git`

if (remote) {
  tempDir = mkdtempSync(join(tmpdir(), "bb-branch-code-"))
  gitDir = tempDir
  ref = "FETCH_HEAD"
  console.error(`Fetching ${branch} from ${remote} into a temporary repo...`)
  const init = spawnSync("git", ["-C", gitDir, "init", "--quiet"], { encoding: "utf8" })
  if (init.status !== 0) fail(output(init), init.status || 1, { remote })
  const fetch = spawnSync("git", ["-C", gitDir, "fetch", "--depth=1", remote, branch], { encoding: "utf8" })
  if (fetch.status !== 0) fail(output(fetch), fetch.status || 1, { remote })
} else {
  gitDir = "."
  console.error(`Fetching origin/${branch}...`)
  const fetch = spawnSync("git", ["fetch", "origin", branch], { encoding: "utf8" })
  if (fetch.status !== 0) fail(output(fetch), fetch.status || 1)
}

if (!filePath) {
  console.error(`Listing files in ${ref}...`)
  const files = spawnSync("git", ["-C", gitDir, "ls-tree", "-r", "--name-only", ref], { encoding: "utf8" })
  if (files.status !== 0) {
    fail(output(files), files.status || 1, { workspace: workspace || null, repo: repo || null, remote: remote || null })
  }
  console.log(JSON.stringify({ ok: true, command: "git ls-tree", branch, ref, workspace: workspace || null, repo: repo || null, remote: remote || null, files: files.stdout.trim().split("\n").filter(Boolean) }, null, 2))
  process.exit(0)
}

console.error(`Reading ${filePath} from ${ref}...`)

const show = spawnSync("git", ["-C", gitDir, "show", `${ref}:${filePath}`], { encoding: "utf8", maxBuffer: 20 * 1024 * 1024 })
if (show.status !== 0) {
  fail(output(show), show.status || 1, { workspace: workspace || null, repo: repo || null, remote: remote || null, path: filePath })
}

console.log(JSON.stringify({ ok: true, command: "git show", branch, ref, workspace: workspace || null, repo: repo || null, remote: remote || null, path: filePath, content: show.stdout }, null, 2))
