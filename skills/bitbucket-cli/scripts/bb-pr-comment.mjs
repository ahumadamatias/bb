#!/usr/bin/env node

import { spawnSync } from "node:child_process"

const argsIn = process.argv.slice(2)
const task = argsIn.includes("--task")
const args = argsIn.filter((arg) => arg !== "--task")
const [id, body, path, line] = args

if (!id || !/^\d+$/.test(id) || !body) {
  console.log(JSON.stringify({ ok: false, command: "bb pr comment", error: "usage: node scripts/bb-pr-comment.mjs <id> <body> [path] [line] [--task]" }, null, 2))
  process.exit(2)
}
if (line && !path) {
  console.log(JSON.stringify({ ok: false, command: "bb pr comment", id: Number(id), error: "line requires path" }, null, 2))
  process.exit(2)
}

console.error(`Commenting on pull request #${id}${task ? " as a task" : ""}...`)

const bbArgs = ["pr", "comment", id, "--body", body]
if (path) bbArgs.push("--path", path)
if (line) bbArgs.push("--line", line)
if (task) bbArgs.push("--task")

const result = spawnSync("bb", bbArgs, { encoding: "utf8" })
if (result.status !== 0) {
  console.log(JSON.stringify({ ok: false, command: "bb pr comment", id: Number(id), error: `${result.stdout || ""}${result.stderr || ""}`.trim() }, null, 2))
  process.exit(result.status || 1)
}

console.log(JSON.stringify({ ok: true, command: "bb pr comment", id: Number(id), path: path || null, line: line ? Number(line) : null, task, output: result.stdout.trim() }, null, 2))
