export const commentPrefix = '  //'

export const staticKeywords = ['module', 'export']

export const declTypes = ['config', 'data', 'database', 'enum', 'fsm', 'topic', 'typealias', 'secret', 'subscription', 'verb']

export const specialChars = ['{', '}', '=']

export function isFirstLineOfBlock(ll: string[], i: number): boolean {
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
