import type { UseQueryResult } from '@tanstack/react-query'
import type { GetModulesResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'

export const commentPrefix = '  //'

export const staticKeywords = ['module', 'export']

export const declTypes = ['config', 'data', 'database', 'enum', 'fsm', 'topic', 'typealias', 'secret', 'subscription', 'verb']

export const declTypeMultiselectOpts = [
  {
    key: 'config',
    displayName: 'Config',
  },
  {
    key: 'data',
    displayName: 'Data',
  },
  {
    key: 'database',
    displayName: 'Database',
  },
  {
    key: 'enum',
    displayName: 'Enum',
  },
  {
    key: 'fsm',
    displayName: 'FSM',
  },
  {
    key: 'topic',
    displayName: 'Topic',
  },
  {
    key: 'typealias',
    displayName: 'Type Alias',
  },
  {
    key: 'secret',
    displayName: 'Secret',
  },
  {
    key: 'subscription',
    displayName: 'Subscription',
  },
  {
    key: 'verb',
    displayName: 'Verb',
  },
]

// Keep these in sync with backend/schema/module.go#L86-L95
const skipNewLineDeclTypes = ['config', 'secret', 'database', 'topic', 'subscription', 'typealias']
const skipGapAfterTypes: { [key: string]: string[] } = {
  secret: ['config'],
  subscription: ['topic'],
}

export const specialChars = ['{', '}', '=']

export function shouldAddLeadingSpace(lines: string[], i: number): boolean {
  if (!isFirstLineOfBlock(lines, i)) {
    return false
  }

  for (const j in skipNewLineDeclTypes) {
    if (declTypeAndPriorLineMatch(lines, i, skipNewLineDeclTypes[j], skipNewLineDeclTypes[j])) {
      return false
    }
  }

  for (const declType in skipGapAfterTypes) {
    for (const j in skipGapAfterTypes[declType]) {
      if (declTypeAndPriorLineMatch(lines, i, declType, skipGapAfterTypes[declType][j])) {
        return false
      }
    }
  }

  return true
}

function declTypeAndPriorLineMatch(lines: string[], i: number, declType: string, priorDeclType: string): boolean {
  if (i === 0 || lines.length === 1) {
    return false
  }
  return regexForDeclType(declType).exec(lines[i]) !== null && regexForDeclType(priorDeclType).exec(lines[i - 1]) !== null
}

function regexForDeclType(declType: string) {
  return new RegExp(`^  (export )?${declType} \\w+`)
}

function isFirstLineOfBlock(lines: string[], i: number): boolean {
  if (i === 0) {
    // Never add space for the first block
    return false
  }
  if (lines[i].startsWith('    ')) {
    // Never add space for nested lines
    return false
  }
  if (lines[i - 1].startsWith(commentPrefix)) {
    // Prior line is a comment
    return false
  }
  if (lines[i].startsWith(commentPrefix)) {
    return true
  }
  const tokens = lines[i].trim().split(' ')
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
