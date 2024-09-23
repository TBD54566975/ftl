import type { UseQueryResult } from '@tanstack/react-query'
import type { GetModulesResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'

export const commentPrefix = '  //'

export const staticKeywords = ['module', 'export']

export const declTypes = ['config', 'data', 'database', 'enum', 'fsm', 'topic', 'typealias', 'secret', 'subscription', 'verb']

// Keep these in sync with backend/schema/module.go#L86-L95
const skipNewLineDeclTypes = ['config', 'secret', 'database', 'topic', 'subscription']
const skipGapAfterTypes: { [key: string]: string[] } = {
  secret: ['config'],
  subscription: ['topic'],
}

export const specialChars = ['{', '}', '=']

export function shouldAddLeadingSpace(ll: string[], i: number): boolean {
  if (!isFirstLineOfBlock(ll, i)) {
    return false
  }

  for (const j in skipNewLineDeclTypes) {
    if (declTypeAndPriorLineMatch(ll, i, skipNewLineDeclTypes[j], skipNewLineDeclTypes[j])) {
      return false
    }
  }

  for (const declType in skipGapAfterTypes) {
    for (const j in skipGapAfterTypes[declType]) {
      if (declTypeAndPriorLineMatch(ll, i, declType, skipGapAfterTypes[declType][j])) {
        return false
      }
    }
  }

  return true
}

function declTypeAndPriorLineMatch(ll: string[], i: number, declType: string, priorDeclType: string): boolean {
  if (i === 0 || ll.length === 1) {
    return false
  }
  return regexForDeclType(declType).exec(ll[i]) !== null && regexForDeclType(priorDeclType).exec(ll[i - 1]) !== null
}

function regexForDeclType(declType: string) {
  return new RegExp(`^  (export )?${declType} \\w+`)
}

function isFirstLineOfBlock(ll: string[], i: number): boolean {
  if (i === 0) {
    // Never add space for the first block
    return false
  }
  if (ll[i].startsWith('    ')) {
    // Never add space for nested lines
    return false
  }
  if (ll[i - 1].startsWith(commentPrefix)) {
    // Prior line is a comment
    return false
  }
  if (ll[i].startsWith(commentPrefix)) {
    return true
  }
  const tokens = ll[i].trim().split(' ')
  if (!tokens || tokens.length === 0) {
    return false
  }
  return staticKeywords.includes(tokens[0]) || declTypes.includes(tokens[0])
}

export interface DeclSchema {
  schema: string
  declType: string
}

export function declFromModules(moduleName: string, declName: string, modules: UseQueryResult<GetModulesResponse, Error>) {
  if (!modules.isSuccess || modules.data.modules.length === 0) {
    return
  }
  const module = modules.data.modules.find((module) => module.name === moduleName)
  if (!module?.schema) {
    return
  }
  return declFromModuleSchemaString(declName, module.schema)
}

export function declFromModuleSchemaString(declName: string, schema: string) {
  const lines = schema.split('\n')
  const foundIdx = lines.findIndex((line) => {
    const regex = new RegExp(`^  (export )?\\w+ ${declName}`)
    return line.match(regex)
  })

  if (foundIdx === -1) {
    return
  }

  const line = lines[foundIdx]
  let out = line
  let subLineIdx = foundIdx + 1
  while (subLineIdx < lines.length && lines[subLineIdx].startsWith('    ')) {
    out += `\n${lines[subLineIdx]}`
    subLineIdx++
  }
  // Check for closing parens
  if (subLineIdx < lines.length && line.endsWith('{') && lines[subLineIdx] === '  }') {
    out += '\n  }'
  }

  // Scan backwards for comments
  subLineIdx = foundIdx - 1
  while (subLineIdx >= 0 && lines[subLineIdx].startsWith(commentPrefix)) {
    out = `${lines[subLineIdx]}\n${out}}`
    subLineIdx--
  }

  const regexExecd = new RegExp(` (\\w+) ${declName}`).exec(line)
  return {
    schema: out,
    declType: regexExecd ? regexExecd[1] : '',
  }
}
