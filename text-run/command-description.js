const child_process = require("child_process")
const diff = require("jsdiff-console")
const getCommand = require("./helpers/get-command.js")

module.exports = async function(activity) {
  const mdDesc = getMd(activity)
  const cliDesc = getCliDesc(activity)
  diff(mdDesc, cliDesc)
}

function getMd(activity) {
  return normalize(
    activity.nodes
      .map(nodeContent)
      .map(text => text.trim())
      .join("\n")
      .replace(/\n<\/?a>\n/g, " ")
  )
}

function nodeContent(node) {
  if (node.type === "list_item_open") {
    return ":" + node.content
  }
  console.log(node.type)
  if (node.type === "list_close") {
    return node.content + ":"
  }
  return node.content
}

function getCliDesc(activity) {
  const command = getCommand(activity.file)
  const output = child_process.execSync(`git-town help ${command}`).toString()
  const matches = output.match(/^.*\n\n([\s\S]*)\n\nUsage:\n/m)
  const lines = matches[1].split("\n")
  let i = 1
  while (i < lines.length) {
    if (lines[i].startsWith("- ")) {
      lines[i] = lines[i].replace(/^- /, "")
      i += 1
      continue
    }
    const lineMatch = lines[i].match(/[0-9]\. /)
    if (lineMatch) {
      lines[i] = lines[i].replace(lineMatch[0], "")
      i += 1
      continue
    }
    if (isTextLine(lines[i])) {
      console.log(11111111111, i)
      lines[i - 1] += lines[i]
      lines.splice(i, 1)
      console.log(lines)
      continue
    }
    i += 1
  }
  console.log(lines)
  return lines
    .join("\n")
    .replace(/\n\n/g, "\n")
    .replace(/,\s*/g, ",\n")
    .replace(/[ ]+/g, " ")
}

function normalize(text) {
  return text
    .replace(/\n/g, " ")
    .replace(/[ ]+/g, " ")
    .replace(/\s*\./g, ".\n")
    .replace(/\s*\,/g, ",\n")
    .replace(/"/g, "\n")
    .replace(/:/g, "\n")
    .replace(/^\s+/gm, "")
    .replace(/\s+$/gm, "")
    .trim()
}

function isTextLine(line) {
  return /^\s*\w+/.test(line)
}
