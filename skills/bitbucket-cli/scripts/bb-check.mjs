#!/usr/bin/env node

import { spawnSync } from "node:child_process"

function run(command, args) {
  return spawnSync(command, args, { encoding: "utf8" })
}

function text(result) {
  return `${result.stdout || ""}${result.stderr || ""}`.trim()
}

console.error("Checking bb installation and authentication...")

const version = run("bb", ["--version"])
if (version.error?.code === "ENOENT") {
  console.log(JSON.stringify({ ok: false, command: "bb --version", installed: false, error: "bb not found in PATH" }, null, 2))
  process.exit(1)
}
if (version.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb --version", installed: false, error: text(version) }, null, 2))
  process.exit(version.status || 1)
}

const auth = run("bb", ["auth", "status"])
if (auth.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb auth status", installed: true, version: version.stdout.trim(), authenticated: false, error: text(auth) }, null, 2))
  process.exit(auth.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb auth status", installed: true, version: version.stdout.trim(), authenticated: true, status: auth.stdout.trim() }, null, 2))
